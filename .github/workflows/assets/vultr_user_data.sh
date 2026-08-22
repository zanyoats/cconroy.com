#!/bin/sh
set -eu
exec > /var/log/user-data.log 2>&1
echo "user-data start: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
# This script sets up a new user with passwordless SSH login and sudo privileges on Alpine Linux.
# It installs sudo and openssh, creates a new user (defaulting to "charlie" if no username is provided),
# removes the user’s password (so only key authentication is allowed), adds the user to the wheel group,
# enables sudo for the wheel group (uncommenting the appropriate line in /etc/sudoers), and installs your public key for SSH.


# Install core services and cert tooling.
apk add --no-cache sudo ufw openssh rsync curl git ca-certificates acl shadow certbot

# Set username (defaults to "charlie" if no argument is provided)
USER_NAME=${1:-charlie}

# Create the new user with default settings (no password prompt)
adduser -D "$USER_NAME"

# Remove the user's password to enforce passwordless login
passwd -d "$USER_NAME"

# Add the new user to the wheel group for sudo privileges
adduser "$USER_NAME" wheel

## The following line in /etc/sudoers enables passwordless sudo for wheel users:
## # %wheel ALL=(ALL:ALL) NOPASSWD: ALL
# Uncomment it if it is present.
if grep -q "^# *%wheel ALL=(ALL:ALL) NOPASSWD: ALL" /etc/sudoers; then
  sed -i 's/^# *\(%wheel ALL=(ALL:ALL) NOPASSWD: ALL\)/\1/' /etc/sudoers
fi

# Set up SSH authorized_keys for the user
USER_HOME=$(getent passwd "$USER_NAME" | cut -d: -f6)
SSH_DIR="$USER_HOME/.ssh"
mkdir -p "$SSH_DIR"
chmod 700 "$SSH_DIR"
echo "__SSH_PUBLIC_KEY__" > "$SSH_DIR/authorized_keys"
chmod 600 "$SSH_DIR/authorized_keys"
chown -R "$USER_NAME:$USER_NAME" "$SSH_DIR"
echo "User '$USER_NAME' has been created with passwordless login and sudo privileges, and your public key has been added."
rc-service sshd start
rc-update add sshd default

# Set up firewall
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow http
ufw allow https
ufw --force enable
rc-service ufw start
rc-update add ufw default

printf "%s%s%s%s\n" \
    "@nginx " \
    "http://nginx.org/packages/alpine/v" \
    `egrep -o '^[0-9]+\.[0-9]+' /etc/alpine-release` \
    "/main" \
    | tee -a /etc/apk/repositories

curl -o /tmp/nginx_signing.rsa.pub https://nginx.org/keys/nginx_signing.rsa.pub

mv /tmp/nginx_signing.rsa.pub /etc/apk/keys/

apk add nginx@nginx

if ! getent group www-data >/dev/null 2>&1; then
  addgroup -S www-data
fi

setfacl -m u:"${USER_NAME}":rwx /usr/share/nginx/html
usermod -aG www-data "${USER_NAME}"
chgrp -R www-data /usr/share/nginx/html
chmod -R g+w /usr/share/nginx/html

SITE_CONF_B64="__SITE_CONF_B64__"
mkdir -p /etc/nginx/conf.d
SITE_CONF_TEMPLATE="/etc/nginx/site.conf.template"
printf '%s' "${SITE_CONF_B64}" | base64 -d > "${SITE_CONF_TEMPLATE}"
SERVER_NAMES="$(awk '
  /^[[:space:]]*server_name[[:space:]]+/ {
    for (i = 2; i <= NF; i++) {
      gsub(/;/, "", $i)
      if ($i != "") print $i
    }
  }
' "${SITE_CONF_TEMPLATE}" | awk '!seen[$0]++' | xargs)"

if [ -z "${SERVER_NAMES}" ]; then
  SERVER_NAMES="_"
fi

cat > /etc/nginx/conf.d/default.conf <<EOF
server {
    listen 80;
    server_name ${SERVER_NAMES};

    root /usr/share/nginx/html;
    index index.html index.htm;

    location ^~ /.well-known/acme-challenge/ {
        root /usr/share/nginx/html;
        default_type "text/plain";
        allow all;
    }

    location / {
        try_files \$uri \$uri/ =404;
    }
}
EOF

rc-service nginx start
rc-update add nginx default

if [ "${SERVER_NAMES}" != "_" ]; then
  DOMAIN_FLAGS=""
  for domain in ${SERVER_NAMES}; do
    DOMAIN_FLAGS="${DOMAIN_FLAGS} -d ${domain}"
  done

  # shellcheck disable=SC2086
  if certbot certonly --webroot -w /usr/share/nginx/html ${DOMAIN_FLAGS} \
    --non-interactive \
    --agree-tos \
    --register-unsafely-without-email \
    --keep-until-expiring; then
    cp "${SITE_CONF_TEMPLATE}" /etc/nginx/conf.d/default.conf
    nginx -t
    rc-service nginx reload

    cat > /usr/local/bin/renew-certs.sh <<'EOF'
#!/bin/sh
set -eu
certbot renew --quiet --deploy-hook "rc-service nginx reload"
EOF
    chmod 755 /usr/local/bin/renew-certs.sh

    if [ ! -f /etc/crontabs/root ]; then
      touch /etc/crontabs/root
    fi
    if ! grep -Fq "/usr/local/bin/renew-certs.sh" /etc/crontabs/root; then
      echo "19 3 * * * /usr/local/bin/renew-certs.sh >> /var/log/certbot-renew.log 2>&1" >> /etc/crontabs/root
    fi
    rc-service crond start
    rc-update add crond default

    echo "HTTPS configured and renewal enabled."
  else
    echo "Initial cert request failed. Point DNS to this server and rerun:"
    echo "certbot certonly --webroot -w /usr/share/nginx/html ${DOMAIN_FLAGS} --non-interactive --agree-tos --register-unsafely-without-email --keep-until-expiring"
  fi
else
  echo "No server_name domains found in ${SITE_CONF_TEMPLATE}; skipping certificate setup."
fi

echo "user-data ok: $(date -u +%Y-%m-%dT%H:%M:%SZ)" > /var/log/user-data.ok
