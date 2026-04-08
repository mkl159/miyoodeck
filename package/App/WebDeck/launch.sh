#!/bin/sh
# Onion Web Deck - Launch script
# Starts the WebDeck server and shows the IP address on screen

APPDIR=$(dirname "$0")
SYSDIR=/mnt/SDCARD/.tmp_update
WEBDECK_BIN="$APPDIR/webdeck"
WEBDECK_PORT=8080
PIDFILE=/tmp/webdeck.pid
LOGFILE=/tmp/webdeck.log

export LD_LIBRARY_PATH="/lib:/config/lib:/mnt/SDCARD/miyoo/lib:$SYSDIR/lib:$SYSDIR/lib/parasyte"
export PATH="$SYSDIR/bin:$PATH"
export HOME=/mnt/SDCARD

cd "$APPDIR"

# ── Check if already running ─────────────────────────────────────────────────
if [ -f "$PIDFILE" ] && kill -0 "$(cat $PIDFILE)" 2>/dev/null; then
    IP=$(ip route get 1 2>/dev/null | awk '{print $NF;exit}')
    infoPanel \
        -t "Web Deck" \
        -m "Already running!\nhttp://$IP:$WEBDECK_PORT" \
        --timeout 4 &
    exit 0
fi

# ── Ensure WiFi is on ─────────────────────────────────────────────────────────
if ! ifconfig wlan0 >/dev/null 2>&1; then
    infoPanel -t "Web Deck" -m "Turning on WiFi..." --persistent &
    PANEL_PID=$!
    /customer/app/axp_test wifion 2>/dev/null
    sleep 2
    ifconfig wlan0 up 2>/dev/null
    /mnt/SDCARD/miyoo/app/wpa_supplicant -B -D nl80211 -iwlan0 \
        -c /appconfigs/wpa_supplicant.conf 2>/dev/null
    udhcpc -i wlan0 -s /etc/init.d/udhcpc.script >/dev/null 2>&1 &
    sleep 3
    kill $PANEL_PID 2>/dev/null
fi

# ── Wait for IP ───────────────────────────────────────────────────────────────
IP=""
TRIES=0
while [ -z "$IP" ] && [ $TRIES -lt 10 ]; do
    IP=$(ip route get 1 2>/dev/null | awk '{print $NF;exit}')
    [ -z "$IP" ] && sleep 1
    TRIES=$((TRIES + 1))
done

# ── No WiFi → ask user to configure it first ─────────────────────────────────
if [ -z "$IP" ]; then
    infoPanel \
        -t "Web Deck" \
        -m "No WiFi connection found.\nPlease connect to WiFi first\n(Apps > Network settings)" \
        --timeout 5 &
    exit 1
fi

# ── Start the server ──────────────────────────────────────────────────────────
chmod +x "$WEBDECK_BIN"
"$WEBDECK_BIN" >> "$LOGFILE" 2>&1 &
echo $! > "$PIDFILE"

sleep 1

# ── Show IP on screen ─────────────────────────────────────────────────────────
infoPanel \
    -t "Web Deck Started" \
    -m "Open in your browser:\nhttp://$IP:$WEBDECK_PORT\n\nServer running in background.\nPress MENU to return." \
    --timeout 8 &

exit 0
