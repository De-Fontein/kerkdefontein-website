# Server runbook

The site runs on one Hetzner Cloud server (Debian 13). A systemd timer runs `fontein-build` every minute; Caddy
serves `/srv/site/current`. Secrets live only on the server, never in this repository.

## First-time setup

1. **Hetzner Cloud console**:
   - Create a project `fontein` and add your SSH public key.
   - Create a server: Falkenstein or Nuremberg, Debian 13, the smallest x86 shared plan, IPv4 + IPv6.
   - Create a Firewall that allows inbound TCP 22, 80 and 443, plus UDP 443, and attach it to the server.
2. **Local `~/.ssh/config`**:

   ```
   Host fontein
     HostName <server IPv4>
     User deploy
   ```

3. **As root on the server, once**:

   ```bash
   apt update && apt -y full-upgrade
   apt -y install caddy libvips-tools poppler-utils brotli unattended-upgrades
   dpkg-reconfigure -plow unattended-upgrades
   adduser --disabled-password --gecos "" deploy && usermod -aG sudo deploy
   install -d -m 700 -o deploy -g deploy /home/deploy/.ssh
   cp /root/.ssh/authorized_keys /home/deploy/.ssh/ && chown deploy:deploy /home/deploy/.ssh/authorized_keys
   echo 'deploy ALL=(ALL) NOPASSWD: /usr/bin/install, /usr/bin/systemctl' > /etc/sudoers.d/deploy
   sed -i 's/^#\?PasswordAuthentication .*/PasswordAuthentication no/; s/^#\?PermitRootLogin .*/PermitRootLogin no/' /etc/ssh/sshd_config
   systemctl reload ssh
   adduser --system --group --no-create-home fontein
   install -d -o fontein -g fontein -m 755 /srv/site /var/cache/fontein
   install -d -o root -g fontein -m 750 /etc/fontein
   ```

4. **Secrets.** Copy them straight to the server:

   ```bash
   scp ~/.config/fontein/sa.json fontein:/tmp/sa.json
   ssh fontein 'sudo install -o root -g fontein -m 640 /tmp/sa.json /etc/fontein/service-account.json && rm /tmp/sa.json'
   ```

   Write `/etc/fontein/config.json` from `config.example.json`, with mode `640` and owner `root:fontein`.
5. **Alerts (optional).** None are used (decision 2026-10-03), so leave `healthcheckUrl` empty. To get an email
   when builds stop: create a healthchecks.io check (period 1 minute, grace 15 minutes) and put its ping URL there.
6. **Units and Caddy**:

   ```bash
   scp deploy/fontein-build.service deploy/fontein-build.timer deploy/Caddyfile fontein:/tmp/
   ssh fontein 'sudo install -m 644 /tmp/fontein-build.service /tmp/fontein-build.timer /etc/systemd/system/ \
     && sudo install -m 644 /tmp/Caddyfile /etc/caddy/Caddyfile && sudo systemctl daemon-reload && sudo systemctl reload caddy'
   ```

## Deploy

```bash
make deploy                                   # build, copy, run once
ssh fontein 'sudo systemctl enable --now fontein-build.timer'   # first deploy only
```

## Check

```bash
ssh fontein 'journalctl -u fontein-build -n 20 --no-pager'   # JSON lines, one run_id per run
```

Each run logs one of the following:
- `no change`
- `published` (with flyer, event and document counts)
- `file skipped` (a volunteer file was rejected, with the reason)
- `source failed, keeping current release`
