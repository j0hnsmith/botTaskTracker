# botTaskTracker Systemd Service

## Service Info
- **Type**: User service (no sudo required)
- **Location**: `~/.config/systemd/user/bottasktracker.service`
- **URL**: http://100.96.20.114:7002 (Tailscale)
- **Logs**: `journalctl --user -u bottasktracker -f`

## Common Commands

```bash
# Restart after code updates
systemctl --user restart bottasktracker

# Check status
systemctl --user status bottasktracker

# View logs (follow)
journalctl --user -u bottasktracker -f

# View recent logs
journalctl --user -u bottasktracker -n 50

# Stop service
systemctl --user stop bottasktracker

# Start service
systemctl --user start bottasktracker
```

## Update Workflow

When updating code:

```bash
cd /home/openclaw/.openclaw/workspace/botTaskTracker

# Pull changes or edit files
git pull  # or make your changes

# Rebuild
make build

# Restart service
systemctl --user restart bottasktracker

# Check it started OK
systemctl --user status bottasktracker
```

## Database Backup

The SQLite database is at: `data/bot_task_tracker.db`

Consider setting up a cron job to back it up:
```bash
# Add to crontab -e
0 2 * * * cp /home/openclaw/.openclaw/workspace/botTaskTracker/data/bot_task_tracker.db /home/openclaw/.openclaw/workspace/botTaskTracker/data/backups/bot_task_tracker_$(date +\%Y\%m\%d).db
```

## Service Details

The service runs as user `openclaw` and:
- Auto-starts on boot (linger enabled)
- Auto-restarts on crash (RestartSec=5)
- Logs to systemd journal
- Persists across logout
