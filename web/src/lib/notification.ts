import { createClient } from "@connectrpc/connect";
import { NotificationService } from "@/gen/proto/notification/v1/notification_pb";
import { transport } from "./connect";

export const notificationClient = createClient(NotificationService, transport);
