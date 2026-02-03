package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/kqnade/vrc-go/vrchat"
	"golang.org/x/term"
)

func main() {
	ctx := context.Background()
	reader := bufio.NewReader(os.Stdin)

	// VRChatクライアントを作成
	client, err := vrchat.NewClient()
	if err != nil {
		log.Fatalf("クライアント作成エラー: %v", err)
	}

	// cookies.jsonが存在するかチェック
	if _, err := os.Stat("cookies.json"); err == nil {
		fmt.Println("cookies.jsonが見つかりました")
		fmt.Print("保存されたクッキーでログインしますか？ (y/n): ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)

		if strings.ToLower(response) == "y" {
			fmt.Println("\nクッキーで認証中...")
			if err := client.LoadCookies("cookies.json"); err != nil {
				fmt.Printf("⚠ クッキーの読み込みに失敗: %v\n", err)
				fmt.Println("手動でログインしてください\n")
				performManualLogin(ctx, reader, client)
			} else {
				// クッキーが有効か確認
				_, err := client.GetCurrentUser(ctx)
				if err != nil {
					fmt.Printf("⚠ クッキーが無効です: %v\n", err)
					fmt.Println("手動でログインしてください\n")
					performManualLogin(ctx, reader, client)
				} else {
					fmt.Println("✓ クッキーでのログインに成功しました")
				}
			}
		} else {
			fmt.Println()
			performManualLogin(ctx, reader, client)
		}
	} else {
		fmt.Println("cookies.jsonが見つかりません")
		fmt.Println("手動でログインしてください\n")
		performManualLogin(ctx, reader, client)
	}

	// ログイン成功後、ユーザー情報を取得
	user, err := client.GetCurrentUser(ctx)
	if err != nil {
		log.Fatalf("ユーザー情報の取得に失敗: %v", err)
	}

	// ログイン成功
	fmt.Println("✓ ログイン成功!")
	fmt.Printf("\nようこそ、%s さん!\n", user.DisplayName)
	fmt.Printf("User ID: %s\n", user.ID)
	fmt.Printf("Status: %s\n", user.Status)
	if user.StatusDescription != "" {
		fmt.Printf("Status Description: %s\n", user.StatusDescription)
	}

	// Cookieを保存
	if err := client.SaveCookies("cookies.json"); err != nil {
		log.Printf("Cookie保存エラー: %v", err)
	} else {
		fmt.Println("✓ Cookieをcookies.jsonに保存しました")
	}
}

func performManualLogin(ctx context.Context, reader *bufio.Reader, client *vrchat.Client) {
	// ユーザー名の入力
	fmt.Print("ユーザー名: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("入力エラー: %v", err)
	}
	username = strings.TrimSpace(username)

	// パスワードの入力（隠す）
	fmt.Print("パスワード: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatalf("パスワード入力エラー: %v", err)
	}
	password := string(passwordBytes)
	fmt.Println() // 改行

	// ログイン試行（まず2FAなしで）
	fmt.Println("\nログイン中...")
	authConfig := vrchat.AuthConfig{
		Username: username,
		Password: password,
	}

	err = client.Authenticate(ctx, authConfig)

	// 2FA認証が必要な場合
	if err != nil && strings.Contains(err.Error(), "two-factor authentication required") {
		fmt.Println("2FA認証が必要です")
		fmt.Print("2FAコード: ")
		code, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("入力エラー: %v", err)
		}
		code = strings.TrimSpace(code)

		// 2FAコード付きで再認証
		authConfig.TOTPCode = code
		if err := client.Authenticate(ctx, authConfig); err != nil {
			log.Fatalf("2FA認証失敗: %v", err)
		}
	} else if err != nil {
		log.Fatalf("ログイン失敗: %v", err)
	}
}
