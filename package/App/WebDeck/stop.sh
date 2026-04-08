#!/bin/sh
# Stop the WebDeck server
PIDFILE=/tmp/webdeck.pid

if [ -f "$PIDFILE" ]; then
    kill "$(cat $PIDFILE)" 2>/dev/null
    rm -f "$PIDFILE"
    echo "Web Deck stopped."
else
    pkill -f webdeck 2>/dev/null
    echo "Web Deck process killed."
fi
