# Remove set -e to make script more robust
# set -e

# Configuration
MOBILE_DIR="../../locolive-mobile"
API_FILE="$MOBILE_DIR/src/services/api.ts"
CHAT_FILE="$MOBILE_DIR/src/hooks/useChat.ts"
LOG_FILE="server.log"
TUNNEL_LOG="tunnel.log"

# Cleanup function
cleanup() {
    echo "Stopping servers..."
    pkill -P $$ # Kill child processes
    kill $BACKEND_PID || true
    pkill -f "localtunnel" || true
    echo "Done."
    exit 0
}
trap cleanup SIGINT SIGTERM

echo "🚀 Starting LocoLiv Development Environment..."

# 0. Cleanup Port 8080
if lsof -i :8080 -t >/dev/null; then
    echo "⚠️  Port 8080 is in use. Killing old process..."
    kill -9 $(lsof -i :8080 -t)
fi

# 1. Start Backend
echo "Starting Go Backend..."
go run cmd/server/main.go > "$LOG_FILE" 2>&1 &
BACKEND_PID=$!
echo "Backend started (PID: $BACKEND_PID). Logs: $LOG_FILE"

# Wait for backend to be ready
echo "Waiting for backend to start..."
sleep 5

# 2. Start Tunnel
echo "Starting Tunnel..."
npx -y localtunnel --port 8080 > "$TUNNEL_LOG" 2>&1 &
TUNNEL_PID=$!

# Wait for tunnel URL
echo "Waiting for tunnel URL..."
sleep 5
while ! grep -q "your url is:" "$TUNNEL_LOG"; do
  sleep 1
done

# Extract URL
TUNNEL_URL=$(grep "your url is:" "$TUNNEL_LOG" | awk '{print $4}')
echo "✅ Tunnel Live: $TUNNEL_URL"

# 3. Update Frontend Code
echo "Updating Mobile App Config..."

# Update config.ts
CONFIG_FILE="$MOBILE_DIR/src/config.ts"
# We use sed to replace the line starting with 'export const API_URL ='
# Note: Initial file has single quotes, we keep them.
sed -i '' "s|export const API_URL = '.*';|export const API_URL = '$TUNNEL_URL';|" "$CONFIG_FILE"

echo "✅ App Config Updated!"
echo "   - API_URL set to: $TUNNEL_URL"

echo ""
echo "🎉 Environment Ready! Press Ctrl+C to stop."
echo "   Backend Logs: tail -f $LOG_FILE"
echo ""

# Keep script running
echo "Press Ctrl+C to stop."
wait $BACKEND_PID
