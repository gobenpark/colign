import { createClient } from "@connectrpc/connect";
import { MemoryService } from "@/gen/proto/memory/v1/memory_pb";
import { transport } from "./connect";

export const memoryClient = createClient(MemoryService, transport);
