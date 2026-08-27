#!/bin/sh

set -eu

umask 077

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
project_dir=$(dirname "$script_dir")
env_path="$project_dir/.env"

show_usage() {
  printf 'Usage: %s [--cloudflare-tunnel] [hostname-or-ip]\n' "$0"
  printf 'Example: %s --cloudflare-tunnel serverdock.example.com\n' "$0"
}

use_cloudflare_tunnel=false
host_was_supplied=false

if [ "${1:-}" = '--cloudflare-tunnel' ]; then
  use_cloudflare_tunnel=true
  shift
fi

case "${1:-}" in
  -h|--help)
    show_usage
    exit 0
    ;;
esac

if [ "$#" -gt 1 ]; then
  show_usage >&2
  exit 1
fi

if [ "$#" -eq 1 ]; then
  tls_host=$1
  host_was_supplied=true
else
  tls_host=${SERVERDOCK_HOST:-localhost}
fi

case "$tls_host" in
  ''|*[!A-Za-z0-9.:-]*)
    printf 'Invalid hostname or IP address: %s\n' "$tls_host" >&2
    exit 1
    ;;
esac

for command_name in docker openssl; do
  if ! command -v "$command_name" >/dev/null 2>&1; then
    printf 'Required command not found: %s\n' "$command_name" >&2
    exit 1
  fi
done

if ! docker compose version >/dev/null 2>&1; then
  printf 'Docker Compose is not available. Install the Docker Compose plugin first.\n' >&2
  exit 1
fi

env_value() {
  key=$1
  [ -f "$env_path" ] || return 0
  awk -v key="$key" 'index($0, key "=") == 1 { print substr($0, length(key) + 2); exit }' "$env_path"
}

set_env_value() {
  sd_env_key=$1
  sd_env_value=$2
  sd_env_tmp_path=$(mktemp "${env_path}.tmp.XXXXXX")

  awk -v key="$sd_env_key" -v value="$sd_env_value" '
    BEGIN { updated = 0 }
    index($0, key "=") == 1 {
      if (!updated) {
        print key "=" value
        updated = 1
      }
      next
    }
    { print }
    END {
      if (!updated) {
        print key "=" value
      }
    }
  ' "$env_path" > "$sd_env_tmp_path"

  chmod 600 "$sd_env_tmp_path"
  mv "$sd_env_tmp_path" "$env_path"
}

public_url_for() {
  sd_public_host=$1
  sd_public_port=$2

  case "$sd_public_host" in
    *:*) sd_public_host="[$sd_public_host]" ;;
  esac

  if [ "$sd_public_port" = '443' ]; then
    printf 'https://%s\n' "$sd_public_host"
  else
    printf 'https://%s:%s\n' "$sd_public_host" "$sd_public_port"
  fi
}

prompt_for_tunnel_token() {
  if [ ! -t 0 ]; then
    printf 'CLOUDFLARE_TUNNEL_TOKEN is required when --cloudflare-tunnel is used non-interactively.\n' >&2
    exit 1
  fi

  printf 'Cloudflare Tunnel token (input is hidden): ' >&2
  saved_terminal_state=$(stty -g)
  trap 'stty "$saved_terminal_state" 2>/dev/null || true' 0 HUP INT TERM
  stty -echo
  if ! IFS= read -r tunnel_token; then
    stty "$saved_terminal_state"
    trap - 0 HUP INT TERM
    printf '\nUnable to read the Cloudflare Tunnel token.\n' >&2
    exit 1
  fi
  stty "$saved_terminal_state"
  trap - 0 HUP INT TERM
  printf '\n' >&2

  if [ -z "$tunnel_token" ]; then
    printf 'Cloudflare Tunnel token cannot be empty.\n' >&2
    exit 1
  fi
}

absolute_path() {
  case "$1" in
    /*) printf '%s\n' "$1" ;;
    *) printf '%s/%s\n' "$project_dir" "${1#./}" ;;
  esac
}

generate_certificate() {
  cert_path=$1
  key_path=$2

  if [ -f "$cert_path" ] && [ -f "$key_path" ]; then
    printf 'Reusing existing TLS certificate and private key.\n'
    return
  fi

  if [ -e "$cert_path" ] || [ -e "$key_path" ]; then
    printf 'TLS certificate and private key must either both exist or both be absent.\n' >&2
    exit 1
  fi

  case "$tls_host" in
    *:*) subject_alt_name="IP:$tls_host" ;;
    *)
      if printf '%s\n' "$tls_host" | grep -Eq '^[0-9]{1,3}(\.[0-9]{1,3}){3}$'; then
        subject_alt_name="IP:$tls_host"
      else
        subject_alt_name="DNS:$tls_host"
      fi
      ;;
  esac

  if [ "$tls_host" = 'localhost' ]; then
    subject_alt_name="$subject_alt_name,IP:127.0.0.1"
  fi

  mkdir -p "$(dirname "$cert_path")" "$(dirname "$key_path")"
  openssl req \
    -x509 \
    -nodes \
    -newkey rsa:3072 \
    -sha256 \
    -days 365 \
    -keyout "$key_path" \
    -out "$cert_path" \
    -subj "/CN=$tls_host" \
    -addext "subjectAltName=$subject_alt_name" \
    -addext 'keyUsage=critical,digitalSignature,keyEncipherment' \
    -addext 'extendedKeyUsage=serverAuth'

  chmod 600 "$key_path"
  chmod 644 "$cert_path"
  printf 'Generated self-signed TLS certificate: %s\n' "$cert_path"
}

configured_tunnel_token=$(env_value CLOUDFLARE_TUNNEL_TOKEN)
tunnel_token=${CLOUDFLARE_TUNNEL_TOKEN:-$configured_tunnel_token}

if [ "$use_cloudflare_tunnel" = true ] && [ -z "$tunnel_token" ]; then
  prompt_for_tunnel_token
fi

if [ -n "$tunnel_token" ]; then
  use_cloudflare_tunnel=true
fi

env_created=false
initial_admin_password='admin123'

if [ ! -f "$env_path" ]; then
  secret_key=$(openssl rand -hex 32)
  ssh_credential_key=$(openssl rand -hex 16)
  https_port=${HTTPS_PORT:-${PORT:-443}}
  tls_cert_path=${TLS_CERT_PATH:-./certs/serverdock.crt}
  tls_key_path=${TLS_KEY_PATH:-./certs/serverdock.key}

  case "$https_port" in
    ''|*[!0-9]*)
      printf 'HTTPS_PORT must be numeric.\n' >&2
      exit 1
      ;;
  esac

  if [ "$https_port" -lt 1 ] || [ "$https_port" -gt 65535 ]; then
    printf 'HTTPS_PORT must be between 1 and 65535.\n' >&2
    exit 1
  fi

  public_url_port=$https_port
  if [ "$use_cloudflare_tunnel" = true ]; then
    public_url_port=443
  fi
  public_url=$(public_url_for "$tls_host" "$public_url_port")

  {
    printf '# Generated by scripts/deploy.sh. Keep this file private and do not delete it.\n'
    printf 'SECRET_KEY=%s\n' "$secret_key"
    printf 'SSH_CREDENTIAL_KEY=%s\n' "$ssh_credential_key"
    printf 'PUBLIC_URL=%s\n' "$public_url"
    printf 'TLS_CERT_PATH=%s\n' "$tls_cert_path"
    printf 'TLS_KEY_PATH=%s\n' "$tls_key_path"
    printf 'DEFAULT_ADMIN_USERNAME=admin\n'
    printf 'DEFAULT_ADMIN_PASSWORD=%s\n' "$initial_admin_password"
    printf 'DEBUG=false\n'
    printf 'HTTPS_PORT=%s\n' "$https_port"
    printf 'DEV_HTTPS_PORT=3000\n'
    if [ "$use_cloudflare_tunnel" = true ]; then
      printf 'CLOUDFLARE_TUNNEL_TOKEN=%s\n' "$tunnel_token"
    fi
  } > "$env_path"
  chmod 600 "$env_path"
  env_created=true
  printf 'Generated secure configuration: %s\n' "$env_path"
else
  chmod 600 "$env_path"

  if [ "$use_cloudflare_tunnel" = true ]; then
    set_env_value CLOUDFLARE_TUNNEL_TOKEN "$tunnel_token"
  fi

  if [ "$host_was_supplied" = true ]; then
    configured_https_port=${HTTPS_PORT:-$(env_value HTTPS_PORT)}
    configured_https_port=${configured_https_port:-${PORT:-$(env_value PORT)}}
    configured_https_port=${configured_https_port:-443}
    public_url_port=$configured_https_port
    if [ "$use_cloudflare_tunnel" = true ]; then
      public_url_port=443
    fi
    set_env_value PUBLIC_URL "$(public_url_for "$tls_host" "$public_url_port")"
  fi

  printf 'Reusing existing configuration and secrets: %s\n' "$env_path"
fi

tls_cert_path=${TLS_CERT_PATH:-$(env_value TLS_CERT_PATH)}
tls_key_path=${TLS_KEY_PATH:-$(env_value TLS_KEY_PATH)}
tls_cert_path=${tls_cert_path:-./certs/serverdock.crt}
tls_key_path=${tls_key_path:-./certs/serverdock.key}
tls_cert_absolute=$(absolute_path "$tls_cert_path")
tls_key_absolute=$(absolute_path "$tls_key_path")

certificate_generated=false
if [ ! -f "$tls_cert_absolute" ] && [ ! -f "$tls_key_absolute" ]; then
  certificate_generated=true
fi

generate_certificate "$tls_cert_absolute" "$tls_key_absolute"

cd "$project_dir"
if [ "$use_cloudflare_tunnel" = true ]; then
  docker compose --env-file "$env_path" --profile cloudflare up --detach --build --wait --wait-timeout 120
else
  docker compose --env-file "$env_path" up --detach --build --wait --wait-timeout 120
fi

public_url=${PUBLIC_URL:-$(env_value PUBLIC_URL)}
https_port=${HTTPS_PORT:-$(env_value HTTPS_PORT)}
https_port=${https_port:-${PORT:-$(env_value PORT)}}
https_port=${https_port:-443}
if [ -z "$public_url" ]; then
  public_url_port=$https_port
  if [ "$use_cloudflare_tunnel" = true ]; then
    public_url_port=443
  fi
  public_url=$(public_url_for "$tls_host" "$public_url_port")
fi

printf '\nServerDock deployed successfully.\n'
printf 'URL:      %s\n' "$public_url"
printf 'Username: %s\n' "${DEFAULT_ADMIN_USERNAME:-$(env_value DEFAULT_ADMIN_USERNAME)}"
if [ "$env_created" = true ]; then
  printf 'Password: %s\n' "$initial_admin_password"
  printf 'Change the initial administrator password immediately after signing in.\n'
else
  printf 'Password: stored in %s\n' "$env_path"
fi

if [ "$certificate_generated" = true ]; then
  if [ "$use_cloudflare_tunnel" = true ]; then
    printf '\nCloudflare manages the browser-facing certificate. The generated self-signed certificate is only used for direct HTTPS access.\n'
  else
    printf '\nThe generated certificate is self-signed. Trust %s on client devices, or replace it with a CA-issued certificate.\n' "$tls_cert_path"
  fi
fi

if [ "$use_cloudflare_tunnel" = true ]; then
  printf 'Cloudflare Tunnel is enabled. Its public hostname must route to http://frontend:80.\n'
  printf 'This publishes the login page to the Internet; change the initial administrator password immediately.\n'
fi
