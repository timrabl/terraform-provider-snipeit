#!/usr/bin/env bash
#
# ci-seed-snipeit.sh - headless seed for a fresh Snipe-IT instance.
#
# A fresh Snipe-IT locks its entire API behind a first-run setup wizard: every
# request (the /api/v1 routes included) redirects to /setup until a settings
# row and an admin user exist. This script bypasses the wizard non-interactively
# so acceptance tests can talk to the API:
#
#   1. wait for the app container to answer /setup
#   2. run migrations and install Passport
#   3. heal storage/cache permissions (the apache worker user differs by version)
#   4. create a superuser (superuser flag forced via a raw DB update, the model guards it)
#   5. insert a settings row if none exists (satisfies the setup gate)
#   6. mint a Passport personal access token
#   7. verify /api/v1/users/me returns JSON before printing the token
#
# On success it prints exactly one line to stdout:  TOKEN=<token>
# All diagnostics go to stderr so the caller can capture the token cleanly.
#
# Usage: ci-seed-snipeit.sh <compose-project-name> <host-port>
#   compose-project-name  the `docker compose -p` value; the app container is
#                         expected at <project>-app-1
#   host-port             the localhost port the app is published on
#
# No secrets are embedded; the admin password is generated per run and only the
# resulting API token leaves the script.

set -euo pipefail

# --- args --------------------------------------------------------------------
PROJECT="${1:-}"
PORT="${2:-}"
if [ -z "$PROJECT" ] || [ -z "$PORT" ]; then
  echo "usage: $0 <compose-project-name> <host-port>" >&2
  exit 2
fi

APP="${PROJECT}-app-1"
BASE="http://localhost:${PORT}"

# Admin credentials are ephemeral: the password never leaves this process and
# the username is fixed so re-runs are idempotent.
ADMIN_USER="ci-admin"
ADMIN_EMAIL="ci-admin@example.test"
ADMIN_PASS="$(openssl rand -base64 24)"

log()  { echo "[seed] $*" >&2; }
fail() { echo "[seed] ERROR: $*" >&2; exit 1; }

# artisan runs a php artisan sub-command inside the app container, non-interactive.
artisan() { docker exec "$APP" php artisan --no-interaction "$@"; }

# tinker pipes a PHP snippet into artisan tinker via stdin. --execute is avoided
# because quoting multi-statement snippets through docker exec is brittle.
tinker() { docker exec -i "$APP" php artisan tinker; }

require_container() {
  if ! docker ps --format '{{.Names}}' | grep -qx "$APP"; then
    log "containers currently running:"
    docker ps --format '  {{.Names}} ({{.Status}})' >&2 || true
    fail "app container '$APP' is not running"
  fi
}

# --- 1. wait for the app to answer -------------------------------------------
# A cold Snipe-IT container spends a while composing autoload files and running
# its own entrypoint before Apache serves anything. Poll /setup (the wizard
# lands there with 200 or 302) until it responds or we time out.
wait_for_app() {
  local deadline=$(( $(date +%s) + 300 ))
  local code
  log "waiting for ${BASE}/setup to answer ..."
  while [ "$(date +%s)" -lt "$deadline" ]; do
    code="$(curl -s -o /dev/null -w '%{http_code}' --max-time 10 "${BASE}/setup" || true)"
    case "$code" in
      200|301|302)
        log "app responded to /setup (HTTP $code)"
        return 0
        ;;
    esac
    sleep 5
  done
  log "last container logs:"
  docker logs --tail 40 "$APP" >&2 || true
  fail "timed out waiting for the app to answer on ${BASE}/setup"
}

# --- 2. migrate + Passport ---------------------------------------------------
migrate_and_passport() {
  log "running migrations ..."
  local i
  for i in 1 2 3; do
    if artisan migrate --force; then
      break
    fi
    log "migrate attempt $i failed, retrying in 10s ..."
    sleep 10
    [ "$i" = 3 ] && fail "migrations did not complete"
  done

  # Passport needs two things for token auth: the RSA key pair in storage/ (used
  # to sign and, crucially, to VALIDATE the bearer token in the auth:api guard)
  # and a personal-access client for createToken(). passport:install is meant to
  # do both, but across Snipe-IT/Passport versions it can return success without
  # writing the storage keys, after which every authenticated request 500s with
  # "oauth-public.key does not exist". So generate the keys explicitly first,
  # then create the personal client, and verify the keys landed.
  log "generating Passport keys ..."
  artisan passport:keys --force || fail "passport:keys failed"

  log "creating Passport personal access client ..."
  # passport:install also creates clients; fall back to a direct client create
  # if it is unhappy on this version.
  artisan passport:install --force >/dev/null 2>&1 || \
    artisan passport:client --personal --name="ci-personal" || \
    fail "could not create a Passport personal access client"

  if ! docker exec "$APP" test -r storage/oauth-public.key; then
    fail "Passport public key is missing after key generation"
  fi
  log "Passport keys present"
}

# --- restart the app so workers pick up the freshly written keys -------------
# The apache workers boot before the Passport keys exist and cache that absence
# in PHP's realpath cache, so at request time the auth guard still reports
# "oauth-public.key does not exist" even though the file is now there. Restart
# the container so fresh workers stat the key for the first time. The DB lives in
# a separate container and the storage/ keys are on the app container's writable
# layer, so a restart preserves both.
restart_app() {
  log "restarting the app container so workers see the new keys ..."
  docker restart "$APP" >/dev/null || fail "could not restart the app container"
  wait_for_app
}

# --- heal storage/cache permissions ------------------------------------------
# The apache worker runs as `docker` on 8.x images and `www-data` on older ones,
# while the master process (and our `docker exec` artisan calls) run as root.
# Anything root writes into storage/ - laravel.log above all - then blocks the
# non-root worker: a request that cannot append to the log 500s with a bare
# "Server Error". Rather than guess the worker user, make the writable trees
# world-writable. This is a disposable, per-run CI instance, so the blunt fix is
# the safe one. Detect the worker user too and hand it ownership as a belt.
heal_permissions() {
  local owner
  owner="$(docker exec "$APP" sh -c \
    "ps axo user,comm 2>/dev/null | awk '\$1 != \"root\" && /apache2|httpd|php-fpm/ {print \$1; exit}'" \
    2>/dev/null || true)"
  [ -z "$owner" ] && owner="www-data"
  log "healing storage permissions (worker user '$owner')"
  docker exec "$APP" sh -c \
    "chown -R '$owner' storage bootstrap/cache 2>/dev/null; \
     chmod -R 0777 storage bootstrap/cache 2>/dev/null; true" || \
    log "warning: could not fully adjust permissions (continuing)"
}

# --- 4-6. create admin, settings row, mint token -----------------------------
seed_admin_and_token() {
  log "creating superuser, settings row and API token ..."
  # One tinker session does the guarded work:
  #   * firstOrNew the admin so re-runs are idempotent
  #   * force the superuser permission via a RAW update (the model $guards it,
  #     so a normal save silently drops it)
  #   * insert a settings row via the query builder if the table is empty (an
  #     Eloquent Setting save is swallowed by the model's singleton/guard logic),
  #     restricted to columns that actually exist so it works across versions
  #   * mint a Passport personal access token and print it with a stable marker
  local out
  out="$(tinker <<PHP 2>>/tmp/seed-tinker.err || true
\$user = \App\Models\User::firstOrNew(['username' => '${ADMIN_USER}']);
\$user->first_name = 'CI';
\$user->last_name  = 'Admin';
\$user->email      = '${ADMIN_EMAIL}';
\$user->password   = \Illuminate\Support\Facades\Hash::make('${ADMIN_PASS}');
\$user->activated  = 1;
\$user->save();

// The superuser flag lives in the guarded permissions column; set it raw.
\DB::table('users')->where('id', \$user->id)
  ->update(['permissions' => json_encode(['superuser' => '1'])]);

// Satisfy the setup gate: at least one settings row must exist.
if (\DB::table('settings')->count() === 0) {
    \$cols = \Illuminate\Support\Facades\Schema::getColumnListing('settings');
    \$row  = array_intersect_key([
        'user_id'    => \$user->id,
        'site_name'  => 'CI',
        'per_page'   => 50,
        'created_at' => now(),
        'updated_at' => now(),
    ], array_flip(\$cols));
    \DB::table('settings')->insert(\$row);
}

\$token = \$user->createToken('ci')->accessToken;
echo 'SEED_TOKEN_START' . \$token . 'SEED_TOKEN_END' . PHP_EOL;
PHP
)"

  TOKEN="$(printf '%s' "$out" | sed -n 's/.*SEED_TOKEN_START\(.*\)SEED_TOKEN_END.*/\1/p' | head -1)"
  if [ -z "$TOKEN" ]; then
    log "tinker output was:"
    printf '%s\n' "$out" >&2
    [ -f /tmp/seed-tinker.err ] && { log "tinker stderr:"; cat /tmp/seed-tinker.err >&2; }
    fail "did not obtain an API token from tinker"
  fi
  log "minted API token (${#TOKEN} chars)"
}

# --- 7. verify the API actually answers with JSON ----------------------------
# The whole point: if the setup gate is still up, /api/v1/users/me returns an
# HTML 302 to /setup instead of JSON. Confirm a real JSON body before we hand
# the token back, retrying briefly to absorb cache warm-up.
verify_api() {
  local i body ctype
  for i in $(seq 1 12); do
    ctype="$(curl -s -o /dev/null -w '%{content_type}' \
      -H "Authorization: Bearer ${TOKEN}" -H 'Accept: application/json' \
      "${BASE}/api/v1/users/me" || true)"
    body="$(curl -s \
      -H "Authorization: Bearer ${TOKEN}" -H 'Accept: application/json' \
      "${BASE}/api/v1/users/me" || true)"
    case "$ctype" in
      application/json*)
        # A valid session returns the user object; guard against a JSON error too.
        if printf '%s' "$body" | grep -q '"id"'; then
          log "verified: /api/v1/users/me returned JSON"
          return 0
        fi
        log "attempt $i: JSON received but not a user object: ${body:0:200}"
        ;;
      *)
        log "attempt $i: non-JSON response (content-type='${ctype:-none}'), retrying ..."
        ;;
    esac
    sleep 5
  done
  log "final response body: ${body:-<empty>}"
  log "application error summary:"
  docker exec "$APP" sh -c \
    'grep -a "production.ERROR" storage/logs/laravel.log 2>/dev/null | tail -3 | cut -c1-500 || true' >&2 || true
  fail "API never returned a JSON user object; the setup gate is likely still up"
}

# --- run ---------------------------------------------------------------------
main() {
  require_container
  wait_for_app
  migrate_and_passport
  # Heal once so the Passport keys just written are readable by the worker, then
  # restart so the workers actually pick them up (realpath cache), then heal
  # again after tinker, whose root-owned laravel.log/cache files are the usual
  # cause of a post-auth 500.
  heal_permissions
  restart_app
  seed_admin_and_token
  heal_permissions
  verify_api
  # The one line the workflow captures.
  echo "TOKEN=${TOKEN}"
}

main "$@"
