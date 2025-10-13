<script setup lang="ts">
import * as z from "zod";
import type { FormSubmitEvent } from "@nuxt/ui";

definePageMeta({
  layout: "auth",
});

useSeoMeta({
  title: "Login",
  description: "Login to your account to continue",
});

const fields = [
  {
    name: "username",
    type: "text" as const,
    label: "Username",
    required: true,
  },
  {
    name: "password",
    label: "Password",
    type: "password" as const,
    placeholder: "Enter your password",
  },
];

const schema = z.object({
  username: z.string(),
  password: z.string(),
});

type Schema = z.output<typeof schema>;

const { loginMutation } = useAuth();

function onSubmit(payload: FormSubmitEvent<Schema>) {
  loginMutation.mutate({
    username: payload.data.username,
    password: payload.data.password,
  });
}
</script>

<template>
  <UAuthForm
    :fields="fields"
    :schema="schema"
    title="Welcome back"
    icon="i-lucide-lock"
    :loading="loginMutation.isPending.value"
    @submit="onSubmit"
  >
    <template v-if="loginMutation.error.value" #validation>
      <UAlert
        color="error"
        icon="i-lucide-info"
        :title="
          'detail' in loginMutation.error.value
            ? loginMutation.error.value.detail as string
            : 'An error occured during authentication.'
        "
      />
    </template>
  </UAuthForm>
</template>
