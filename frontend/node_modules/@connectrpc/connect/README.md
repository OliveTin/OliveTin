# @connectrpc/connect

Connect is a family of libraries for building type-safe APIs with different languages and platforms.
[@connectrpc/connect](https://www.npmjs.com/package/@connectrpc/connect) brings them to TypeScript,
the web browser, and to Node.js.

With Connect, you define your schema first:

```proto
service ElizaService {
  rpc Say(SayRequest) returns (SayResponse) {}
}
```

And with the magic of code generation, this schema produces servers and clients:

```ts
const answer = await eliza.say({ sentence: "I feel happy." });
console.log(answer);
// {sentence: 'When you feel happy, what do you do?'}
```

Unlike REST, the RPCs you use with Connect are typesafe end to end, but they are
regular HTTP under the hood. You can see all requests in the network inspector,
and you can `curl` them if you want:

```shell
curl \
    --header 'Content-Type: application/json' \
    --data '{"sentence": "I feel happy."}' \
    https://demo.connectrpc.com/connectrpc.eliza.v1.ElizaService/Say
```

With Connect for ECMAScript, you can spin up a service in Node.js and call it
from the web, the terminal, or native mobile clients. Under the hood, it uses
[Protocol Buffers](https://github.com/bufbuild/protobuf-es) for the schema, and
implements RPC (remote procedure calls) with three protocols: The widely available
gRPC and gRPC-web, and Connect's [own protocol](https://connectrpc.com/docs/protocol/),
optimized for the web. This gives you unparalleled interoperability with
full-stack type-safety.

## Get started on the web

Follow our [10 minute tutorial](https://connectrpc.com/docs/web/getting-started) where
we use [Vite](https://vitejs.dev/) and [React](https://reactjs.org/) to create a
web interface for ELIZA.

**React**, **Svelte**, **Vue**, **Next.js** and **Angular** are supported (see [examples](https://github.com/connectrpc/examples-es)),
and we have an expansion pack for [TanStack Query](https://github.com/connectrpc/connect-query-es).
We support all modern web browsers that implement the widely available
[fetch API](https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API)
and the [Encoding API](https://developer.mozilla.org/en-US/docs/Web/API/Encoding_API).

## Get started on Node.js

Follow our [10 minute tutorial](https://connectrpc.com/docs/node/getting-started)
to spin up a service in Node.js, and call it from the web, and from a gRPC client
in your terminal.

You can use vanilla Node.js, or our server plugins for [Fastify](https://www.fastify.io/)
or [Express](https://expressjs.com/). We support the builtin `http`, and `http2`
modules on Node.js v18.14.1 and later.
