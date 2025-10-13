import { useQuery } from "@tanstack/vue-query";
import { createGlobalState, useWebSocket } from "@vueuse/core";
import z from "zod";
import { getApiV1Downloads } from "~/utils/client";
import { client } from "~/utils/client/client.gen";

import mitt from "mitt";

const msgSchema = z
  .string()
  .transform((input) => {
    try {
      return JSON.parse(input);
    } catch {
      return null;
    }
  })
  .pipe(
    z.object({
      event: z.enum(["progress", "success", "error", "start"]),
      download_id: z.int(),
      data: z.string(),
    })
  );

const emitter = mitt<{
  [key: `download-${number}`]: z.infer<typeof msgSchema>;
}>();

export const useDownloads = createGlobalState(() => {
  const apiBaseUrl = client.getConfig().baseUrl ?? "";
  const apiHost = apiBaseUrl ? new URL(apiBaseUrl).host : window.location.host;
  const { data: websocketData, close } = useWebSocket(
    `ws://${apiHost}/api/v1/ws/download`
  );

  const downloadsQuery = useQuery({
    queryKey: ["downloads"],
    queryFn: async () => {
      const response = await getApiV1Downloads();
      const downloads = response.data?.downloads ?? [];
      return downloads;
    },
  });

  const parsedWebsocketData = computed(
    () => msgSchema.safeParse(websocketData.value).data
  );

  watch(parsedWebsocketData, (data) => {
    if (data) {
      // emit an event for the specific download_id
      emitter.emit(`download-${data.download_id}`, data);
    }
  });

  return {
    close,
    downloadsQuery,
    parsedWebsocketData,
    websocketEmitter: emitter,
  };
});
