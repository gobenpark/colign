import { Hocuspocus } from "@hocuspocus/server";

const server = new Hocuspocus({
  port: 1234,
  onConnect: async () => {
    console.log("client connected");
  },
});

server.listen();
