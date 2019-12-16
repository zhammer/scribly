create type "public"."email_verification_state" as enum ('pending', 'verified');

alter table "public"."users" add column "email_verification_status" email_verification_state not null
default 'pending'::email_verification_state;
