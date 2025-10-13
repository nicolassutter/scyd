export default defineNuxtRouteMiddleware(async () => {
  const { checkAuthMutation } = useAuth();

  const isAuthenticated = await checkAuthMutation
    .mutateAsync()
    .then((data) => data.authenticated)
    .catch(() => false);

  // redirect the user to the login screen if they're not authenticated
  if (!isAuthenticated) {
    return navigateTo("/login");
  }
});
