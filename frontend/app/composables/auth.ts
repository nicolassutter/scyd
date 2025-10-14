import { useMutation } from "@tanstack/vue-query";
import {
  postApiV1AuthLogin as login,
  postApiV1AuthLogout as logout,
  getApiV1AuthStatus as checkAuth,
} from "../utils/client/sdk.gen";
import { createGlobalState } from "@vueuse/core";

type User = {
  username: string;
};

export const useAuth = createGlobalState(() => {
  const user = ref<User | null>(null);
  const router = useRouter();

  function setUser(usedata: User | null) {
    user.value = usedata;
  }

  const loginMutation = useMutation({
    mutationFn: async ({
      username,
      password,
    }: {
      username: string;
      password: string;
    }) => {
      const res = await login({
        body: { username, password },
      });

      if (!res.data?.success) {
        throw new Error("Login failed");
      }

      return res.data;
    },
    onSuccess: async ({ username }) => {
      setUser({ username });
      await router.push("/");
    },
  });

  const logoutMutation = useMutation({
    mutationFn: async () => {
      await logout();
    },
    onSuccess: async () => {
      user.value = null;
      await router.push("/login");
    },
  });

  const checkAuthMutation = useMutation({
    mutationFn: async (): Promise<{
      authenticated: boolean;
      username?: string;
    }> => {
      if (user.value) {
        return { authenticated: true, username: user.value.username };
      }
      const res = await checkAuth();
      if (!res.data) {
        throw new Error("Failed to check authentication");
      }
      return res.data;
    },
    onSuccess: (data) => {
      if (data?.authenticated && data.username) {
        setUser({ username: data.username });
      } else {
        setUser(null);
      }
    },
  });

  const isAuthenticated = computed(() => user.value !== null);

  return {
    user,
    isAuthenticated,
    loginMutation,
    logoutMutation,
    checkAuthMutation,
  };
});
