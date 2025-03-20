# sway-battery-monitor
Go App for Battery Monitoring in Sway

1. Build the application
```bash
go build .
```

2. Install the binary:
```bash
sudo cp battery-monitor /usr/local/bin
sudo chmod +x /usr/local/bin/battery-monitor
```

3. Set up the systemd user service:
```bash
mkdir -p ~/.config/systemd/user
cp battery-monitor.service ~/.config/systemd/user/
systemctl --user daemon-reload
systemctl --user enable battery-monitor.service
systemctl --user start battery-monitor.service
```
