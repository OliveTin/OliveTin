package websocket

import (
	pb "github.com/OliveTin/OliveTin/gen/grpc"
	"github.com/OliveTin/OliveTin/internal/executor"
	ws "github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"net/http"
	"sync"
)

var upgrader = ws.Upgrader{
	CheckOrigin: checkOriginPermissive,
}

var (
	sendmutex = sync.Mutex{}
)

type WebsocketClient struct {
	conn *ws.Conn
}

var clients []*WebsocketClient

var marshalOptions = protojson.MarshalOptions{
	UseProtoNames:   false, // eg: canExec for js instead of can_exec from protobuf
	EmitUnpopulated: true,
}

var ExecutionListener WebsocketExecutionListener

type WebsocketExecutionListener struct{}

func (WebsocketExecutionListener) OnExecutionStarted(title string) {
	/*
		broadcast(ExecutionStarted{
			Type: "ExecutionStarted",
			Action: title,
		});
	*/
}

func OnEntityChanged() {
	broadcast(&pb.EventEntityChanged{})
}

func (WebsocketExecutionListener) OnActionMapRebuilt() {
	broadcast(&pb.EventConfigChanged{})
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

func (WebsocketExecutionListener) OnOutputChunk(chunk []byte, executionTrackingId string) {
	log.Tracef("outputchunk: %s", string(chunk))

	oc := &pb.EventOutputChunk{
		Output:              string(chunk),
		ExecutionTrackingId: executionTrackingId,
	}

	broadcast(oc)
}

func (WebsocketExecutionListener) OnExecutionFinished(logEntry *executor.InternalLogEntry) {
	evt := &pb.EventExecutionFinished{
		LogEntry: &pb.LogEntry{
			ActionTitle:         logEntry.ActionTitle,
			ActionIcon:          logEntry.ActionIcon,
			ActionId:            logEntry.ActionId,
			DatetimeStarted:     logEntry.DatetimeStarted.Format("2006-01-02 15:04:05"),
			DatetimeFinished:    logEntry.DatetimeFinished.Format("2006-01-02 15:04:05"),
			Stdout:              logEntry.Stdout,
			Stderr:              logEntry.Stderr,
			TimedOut:            logEntry.TimedOut,
			Blocked:             logEntry.Blocked,
			ExitCode:            logEntry.ExitCode,
			Tags:                logEntry.Tags,
			ExecutionTrackingId: logEntry.ExecutionTrackingID,
			ExecutionStarted:    logEntry.ExecutionStarted,
			ExecutionFinished:   logEntry.ExecutionFinished,
		},
	}

	broadcast(evt)
}

func broadcast(pbmsg protoreflect.ProtoMessage) {
	payload, err := marshalOptions.Marshal(pbmsg)

	if err != nil {
		log.Errorf("websocket marshal error: %v", err)
		return
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

	sendmutex.Lock()
	for _, client := range clients {
		client.conn.WriteMessage(ws.TextMessage, hackyMessage)
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

func HandleWebsocket(w http.ResponseWriter, r *http.Request) bool {
	c, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Warnf("Websocket issue: %v", err)
		return false
	}

	//	defer c.Close()

	wsclient := &WebsocketClient{
		conn: c,
	}

	sendmutex.Lock()

	clients = append(clients, wsclient)

	sendmutex.Unlock()

	go wsclient.messageLoop()

	return true
}
