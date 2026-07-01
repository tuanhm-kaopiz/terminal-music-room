package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/spf13/cobra"
	"github.com/terminal-music-room/music-room/internal/client/config"
	"github.com/terminal-music-room/music-room/internal/protocol"
	"github.com/terminal-music-room/music-room/internal/server/room"
)

var (
	loginName   string
	loginServer string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in with a nickname",
	Long:  "Validate nickname, connect to music-roomd, and save session to config.",
	RunE:  runLogin,
}

func init() {
	loginCmd.Flags().StringVar(&loginName, "name", "", "nickname (1–32 characters)")
	loginCmd.Flags().StringVar(&loginServer, "server", "", "music-roomd base URL (default from config, env, or localhost)")
	_ = loginCmd.MarkFlagRequired("name")
	RootCmd.AddCommand(loginCmd)
}

func runLogin(cmd *cobra.Command, _ []string) error {
	nickname, err := room.ValidateNickname(loginName)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	path, err := resolveConfigPath()
	if err != nil {
		return err
	}
	existing, err := config.Load(path)
	if err != nil {
		return err
	}

	serverURL := config.ResolveServerURL(loginServer, existing.ServerURL)
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	ack, err := establishSession(ctx, serverURL, nickname, "")
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	cfg := config.Config{
		Nickname:  nickname,
		ServerURL: serverURL,
		SessionID: ack.SessionID,
	}
	if err := config.Save(path, cfg); err != nil {
		return err
	}

	display := ack.DisplayName
	if display == "" {
		display = nickname
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Logged in as %s\n", display)
	fmt.Fprintf(cmd.OutOrStdout(), "Config saved to %s\n", path)
	return nil
}

func resolveConfigPath() (string, error) {
	if configPath != "" {
		return configPath, nil
	}
	return config.ResolvePath()
}

func establishSession(ctx context.Context, serverURL, nickname, sessionID string) (protocol.SessionAckPayload, error) {
	wsURL, err := config.WebSocketURL(serverURL)
	if err != nil {
		return protocol.SessionAckPayload{}, err
	}
	hdr := http.Header{}
	if sessionID != "" {
		hdr.Set("X-Session-Id", sessionID)
	} else {
		hdr.Set("X-Nickname", nickname)
	}
	conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{HTTPHeader: hdr})
	if err != nil {
		return protocol.SessionAckPayload{}, fmt.Errorf("connect to server: %w", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "login done")

	_, data, err := conn.Read(ctx)
	if err != nil {
		return protocol.SessionAckPayload{}, fmt.Errorf("read session ack: %w", err)
	}
	env, ack, err := protocol.DecodePayload[protocol.SessionAckPayload](data)
	if err != nil {
		return protocol.SessionAckPayload{}, fmt.Errorf("decode session ack: %w", err)
	}
	if env.Type != protocol.MsgSessionAck {
		return protocol.SessionAckPayload{}, fmt.Errorf("expected session.ack, got %q", env.Type)
	}
	if ack.SessionID == "" {
		return protocol.SessionAckPayload{}, fmt.Errorf("server returned empty session id")
	}
	return ack, nil
}

// Login validates nickname and saves session — exported for tests.
func Login(ctx context.Context, w io.Writer, cfgPath, nickname, serverURL string) error {
	loginName = nickname
	loginServer = serverURL
	configPath = cfgPath
	loginCmd.SetOut(w)
	loginCmd.SetErr(w)
	return runLogin(loginCmd, nil)
}