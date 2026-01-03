#!/bin/sh
set -eu
exec > /var/log/user-data.log 2>&1
echo "user-data start: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
# This script sets up a new user with passwordless SSH login and sudo privileges on Alpine Linux.
# It installs sudo and openssh, creates a new user (defaulting to "charlie" if no username is provided),
# removes the user’s password (so only key authentication is allowed), adds the user to the wheel group,
# enables sudo for the wheel group (uncommenting the appropriate line in /etc/sudoers), and installs your public key for SSH.


# Install sudo and openssh
apk add --no-cache sudo ufw openssh rsync curl ca-certificates acl shadow

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
printf '%s' "${SITE_CONF_B64}" | base64 -d > /etc/nginx/conf.d/default.conf

rc-service nginx start

rc-update add nginx default

echo "user-data ok: $(date -u +%Y-%m-%dT%H:%M:%SZ)" > /var/log/user-data.ok
