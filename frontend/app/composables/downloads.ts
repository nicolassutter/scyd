import {
  queryOptions,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/vue-query";
import { createGlobalState, useWebSocket } from "@vueuse/core";
import z from "zod";
import {
  deleteApiV1DownloadById,
  getApiV1Downloads,
  postApiV1DownloadCancelById,
  postApiV1SortDownloads,
  type Download,
} from "~/utils/client";
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

export const downloadStateSchema = z.enum([
  "pending",
  "success",
  "progress",
  "error",
]);
export type DownloadState = z.infer<typeof downloadStateSchema>;

const emitter = mitt<{
  [key: `download-${number}`]: z.infer<typeof msgSchema>;
}>();

export const useDownloads = createGlobalState(() => {
  const apiBaseUrl = client.getConfig().baseUrl ?? "";
  const apiHost = apiBaseUrl ? new URL(apiBaseUrl).host : window.location.host;

  const wsProtocol = window.location.protocol === "https:" ? "wss" : "ws";
  const { data: websocketData, close } = useWebSocket(
    `${wsProtocol}://${apiHost}/api/v1/ws/download`
  );

  const downloadsQueryOptions = queryOptions({
    queryKey: ["downloads"],
    queryFn: async () => {
      const response = await getApiV1Downloads();
      const downloads = response.data?.downloads ?? [];
      return downloads;
    },
  });
  const downloadsQuery = useQuery(downloadsQueryOptions);

  const parsedWebsocketData = computed(
    () => msgSchema.safeParse(websocketData.value).data
  );

  watch(parsedWebsocketData, (data) => {
    if (data) {
      // emit an event for the specific download_id
      emitter.emit(`download-${data.download_id}`, data);
    }
  });

  const queryClient = useQueryClient();

  /**
   * Update a download item in the local cache only
   */
  function updateDownloadItemLocal(
    id: number,
    update: Partial<Download> & {
      state?: DownloadState;
    }
  ) {
    queryClient.setQueryData(downloadsQueryOptions.queryKey, (old) => {
      if (!old) return old;

      return old.map((item) =>
        item.id === id ? { ...item, ...update } : item
      );
    });
  }

  const sortDownloadsMutation = useMutation({
    mutationFn: async () => {
      const res = await postApiV1SortDownloads();
      return res.data;
    },
  });

  const useDeleteDownload = () =>
    useMutation({
      mutationFn: async (id: number) => {
        await deleteApiV1DownloadById({
          path: {
            id,
          },
        });
      },
      onSuccess: (_, id) => {
        // Remove the download from the local cache
        queryClient.setQueryData(downloadsQueryOptions.queryKey, (old) => {
          if (!old) return old;
          return old.filter((item) => item.id !== id);
        });
      },
    });

  const useCancelDownload = () =>
    useMutation({
      mutationFn: async (id: number) => {
        await postApiV1DownloadCancelById({
          path: {
            id,
          },
        });
      },
    });

  return {
    close,
    downloadsQuery,
    sortDownloadsMutation,
    parsedWebsocketData,
    websocketEmitter: emitter,
    updateDownloadItemLocal,
    useDeleteDownload,
    useCancelDownload,
  };
});
