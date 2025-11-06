package httpservers

import (
	"net/http"
	"sync"

	apiv1 "github.com/OliveTin/OliveTin/gen/grpc/olivetin/api/v1"
	"github.com/OliveTin/OliveTin/internal/acl"
	"github.com/OliveTin/OliveTin/internal/config"
	"github.com/OliveTin/OliveTin/internal/executor"
	ws "github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var upgrader = ws.Upgrader{
	CheckOrigin: checkOriginPermissive,
}

var (
	sendmutex = sync.Mutex{}
)

type WebsocketClient struct {
	conn              *ws.Conn
	authenticatedUser *acl.AuthenticatedUser
}

var clients map[*WebsocketClient]struct{}

var marshalOptions = protojson.MarshalOptions{
	UseProtoNames:   false, // eg: canExec for js instead of can_exec from protobuf
	EmitUnpopulated: true,
}

var ExecutionListener WebsocketExecutionListener

type WebsocketExecutionListener struct{}

func (WebsocketExecutionListener) OnExecutionStarted(ile *executor.InternalLogEntry, action *config.Action) {
	evt := &apiv1.EventExecutionStarted{
		LogEntry: internalLogEntryToPb(ile),
	}

	for client := range copyOfClients() {
		if acl.IsAllowedLogs(cfg, client.authenticatedUser, action) {
			writeMessageToClient(client, prepareMessage(evt))
		}
	}
}

func OnEntityChanged() {
	broadcast(&apiv1.EventEntityChanged{})
}

func (WebsocketExecutionListener) OnActionMapRebuilt() {
	broadcast(&apiv1.EventConfigChanged{})
}

/*
The default checkOrigin function checks that the origin (browser) matches the
request origin. However in OliveTin we expect many users to deliberately proxy
the connection with reverse proxies.

So, we just permit any origin. After some searching I'm not sure if this exposes
OliveTin to security issues, but it seems probably not. It would be possible to
create a config option like PermitWebsocketConnectionsFrom or something, but
I'd prefer if OliveTin works as much as possible "out of the box".

If this does expose OliveTin to security issues, it will be changed in the
future obviously.
*/
func checkOriginPermissive(r *http.Request) bool {
	return true
}

func (WebsocketExecutionListener) OnOutputChunk(chunk []byte, executionTrackingId string, logEntry *executor.InternalLogEntry, action *config.Action) {
	log.Tracef("outputchunk: %s", string(chunk))

	oc := &apiv1.EventOutputChunk{
		Output:              string(chunk),
		ExecutionTrackingId: executionTrackingId,
	}

	for client := range copyOfClients() {
		if acl.IsAllowedLogs(cfg, client.authenticatedUser, action) {
			writeMessageToClient(client, prepareMessage(oc))
		}
	}
}

func (WebsocketExecutionListener) OnExecutionFinished(logEntry *executor.InternalLogEntry, action *config.Action) {
	evt := &apiv1.EventExecutionFinished{
		LogEntry: internalLogEntryToPb(logEntry),
	}

	for client := range copyOfClients() {
		if acl.IsAllowedLogs(cfg, client.authenticatedUser, action) {
			writeMessageToClient(client, prepareMessage(evt))
		}
	}

	log.Infof("WS Execution finished: %v", evt.LogEntry)
}

func copyOfClients() map[*WebsocketClient]struct{} {
	sendmutex.Lock()
	defer sendmutex.Unlock()

	if clients == nil {
		clients = make(map[*WebsocketClient]struct{})
	}

	copy := make(map[*WebsocketClient]struct{})
	for client := range clients {
		copy[client] = struct{}{}
	}

	return copy
}

func prepareMessage(pbmsg protoreflect.ProtoMessage) []byte {
	payload, err := marshalOptions.Marshal(pbmsg)

	if err != nil {
		log.Errorf("websocket marshal error: %v", err)
		return nil
	}

	messageType := pbmsg.ProtoReflect().Descriptor().FullName()

	// <EVIL>
	// So, the websocket wants to encode messages using the same protomarshaller
	// as the REST API - this gives consistency instead of using encoding/json
	// and allows us to set specific marshalOptions.
	//
	// However, the protomarshaller will marshal the type, but the JavaScript at
	// the other end has no idea what type this object is - as we're just sending
	// it as JSON over the websocket.
	//
	// Therefore, we wrap the nicely marsheled bytes in a hacky JSON string
	// literal and encode that string just with a byte array cast.
	hackyMessageEnvelope := "{\"type\": \"" + messageType + "\", \"payload\": "

	hackyMessage := []byte{}
	hackyMessage = append(hackyMessage, []byte(hackyMessageEnvelope)...)
	hackyMessage = append(hackyMessage, payload...)
	hackyMessage = append(hackyMessage, []byte("}")...)
	// </EVIL>

	return hackyMessage
}

func broadcast(pbmsg protoreflect.ProtoMessage) {
	message := prepareMessage(pbmsg)

	for client := range copyOfClients() {
		writeMessageToClient(client, message)
	}
}

func writeMessageToClient(client *WebsocketClient, message []byte) {
	if message == nil {
		log.Warnf("writeMessageToClient: message is nil")
		return
	}

	sendmutex.Lock()
	if err := client.conn.WriteMessage(ws.TextMessage, message); err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"client": client,
		}).Debugf("websocket send error")
		_ = client.conn.Close()
		delete(clients, client)
	}
	sendmutex.Unlock()
}

func (c *WebsocketClient) messageLoop() {
	for {
		mt, message, err := c.conn.ReadMessage()

		if err != nil {
			log.Debugf("err: %v", err)
			break
		}

		log.Tracef("websocket recv: %s %d", message, mt)
	}
}

func handleWebsocket(w http.ResponseWriter, r *http.Request) bool {
	c, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Warnf("Websocket issue: %v", err)
		return false
	}

	unauthenticatedUser := authHttpRequest(r)

	authenticatedUser := acl.UserFromUnauthenticatedUser(cfg, unauthenticatedUser)

	wsclient := &WebsocketClient{
		conn:              c,
		authenticatedUser: authenticatedUser,
	}

	sendmutex.Lock()

	if clients == nil {
		clients = make(map[*WebsocketClient]struct{})
	}

	clients[wsclient] = struct{}{}
	sendmutex.Unlock()

	go wsclient.messageLoop()

	return true
}

func internalLogEntryToPb(logEntry *executor.InternalLogEntry) *apiv1.LogEntry {
	return &apiv1.LogEntry{
		ActionTitle:         logEntry.ActionTitle,
		ActionIcon:          logEntry.ActionIcon,
		ActionId:            logEntry.ActionId,
		DatetimeStarted:     logEntry.DatetimeStarted.Format("2006-01-02 15:04:05"),
		DatetimeFinished:    logEntry.DatetimeFinished.Format("2006-01-02 15:04:05"),
		Output:              logEntry.Output,
		TimedOut:            logEntry.TimedOut,
		Blocked:             logEntry.Blocked,
		ExitCode:            logEntry.ExitCode,
		Tags:                logEntry.Tags,
		ExecutionTrackingId: logEntry.ExecutionTrackingID,
		ExecutionStarted:    logEntry.ExecutionStarted,
		ExecutionFinished:   logEntry.ExecutionFinished,
		User:                logEntry.Username,
	}
}
