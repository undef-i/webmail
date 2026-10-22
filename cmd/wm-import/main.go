package main

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"webmail/internal/auth"
	"webmail/internal/mail"
	"webmail/internal/smcrypt"
	"webmail/internal/store"
)

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "失败: "+format+"\n", args...)
	os.Exit(1)
}

func readTTYPassword(prompt string) (string, error) {
	if v := os.Getenv("WM_IMPORT_TEST_PASS"); v != "" {
		return v, nil
	}
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", fmt.Errorf("打开终端: %w", err)
	}
	defer tty.Close()
	fmt.Fprint(tty, prompt)
	off := exec.Command("stty", "-echo")
	off.Stdin = tty
	_ = off.Run()
	defer func() {
		on := exec.Command("stty", "echo")
		on.Stdin = tty
		_ = on.Run()
		fmt.Fprintln(tty)
	}()
	line, err := bufio.NewReader(tty).ReadString('\n')
	if err != nil && line == "" {
		return "", fmt.Errorf("读取输入: %w", err)
	}
	return strings.TrimRight(line, "\r\n"), nil
}

type domainSettings struct {
	IMAP struct {
		Host string `json:"host"`
		Port int    `json:"port"`
		Type int    `json:"type"`
	} `json:"IMAP"`
	SMTP struct {
		Host string `json:"host"`
		Port int    `json:"port"`
		Type int    `json:"type"`
	} `json:"SMTP"`
}

func tlsName(t int) string {
	switch t {
	case 1:
		return "tls"
	case 2:
		return "starttls"
	default:
		return "none"
	}
}

func hostFor(domain, configured string) string {
	if domain == "nedifinita.com" {
		return "mx0.nedifinita.com"
	}
	return configured
}

func loadDomain(smDir, domain string) (domainSettings, error) {
	var ds domainSettings
	raw, err := os.ReadFile(filepath.Join(smDir, "domains", domain+".json"))
	if err != nil {
		return ds, err
	}
	if err := json.Unmarshal(raw, &ds); err != nil {
		return ds, fmt.Errorf("解析 %s 域配置: %w", domain, err)
	}
	return ds, nil
}

func makeAccount(email, username, password string, ds domainSettings, domain string) mail.Account {
	return mail.Account{
		Label:    email,
		Email:    email,
		ImapHost: hostFor(domain, ds.IMAP.Host),
		ImapPort: ds.IMAP.Port,
		ImapTLS:  tlsName(ds.IMAP.Type),
		SMTPHost: hostFor(domain, ds.SMTP.Host),
		SMTPPort: ds.SMTP.Port,
		SMTPTLS:  tlsName(ds.SMTP.Type),
		Username: username,
		Password: password,
	}
}

type addlAccount struct {
	Email string `json:"email"`
	Login string `json:"login"`
	Pass  string `json:"pass"`
	HMAC  string `json:"hmac"`
}

func main() {
	dbPath := flag.String("db", "/var/lib/webmail/webmail.db", "webmail SQLite 数据库路径")
	smDir := flag.String("sm", "/var/www/snappymail/data/_data_/_default_", "SnappyMail _default_ 数据目录")
	keepMaster := flag.Bool("keep-master", false, "不重置主密码：校验当前主密码后只导入新账号")
	flag.Parse()

	dataDir := filepath.Dir(filepath.Dir(*smDir))
	privateName := filepath.Base(*smDir)
	appSalt, err := smcrypt.AppSalt(dataDir, privateName)
	if err != nil {
		fatal("%v", err)
	}

	storageDir := filepath.Join(*smDir, "storage")
	domains, err := os.ReadDir(storageDir)
	if err != nil {
		fatal("读 SnappyMail storage: %v", err)
	}
	var mainDomain, mainUser, mainDir string
	for _, d := range domains {
		if !d.IsDir() {
			continue
		}
		users, _ := os.ReadDir(filepath.Join(storageDir, d.Name()))
		for _, u := range users {
			dir := filepath.Join(storageDir, d.Name(), u.Name())
			if _, err := os.Stat(filepath.Join(dir, ".cryptkey")); err == nil {
				mainDomain, mainUser, mainDir = d.Name(), u.Name(), dir
			}
		}
	}
	if mainDir == "" {
		fatal("在 SnappyMail 数据里找不到带 .cryptkey 的主账号")
	}
	mainEmail := mainUser + "@" + mainDomain
	fmt.Printf("SnappyMail 主账号: %s\n", mainEmail)

	mailPass, err := readTTYPassword("请输入主账号 " + mainEmail + " 的邮箱密码（输入不回显）: ")
	if err != nil {
		fatal("%v", err)
	}
	if mailPass == "" {
		fatal("邮箱密码为空")
	}

	ckRaw, err := os.ReadFile(filepath.Join(mainDir, ".cryptkey"))
	if err != nil {
		fatal("读 .cryptkey: %v", err)
	}
	ckVal, err := smcrypt.DecryptFromJSON(string(ckRaw), []byte(mailPass), appSalt)
	if err != nil {
		fatal("解开 .cryptkey 失败（邮箱密码不对？）: %v", err)
	}
	ckHex, ok := ckVal.(string)
	if !ok || len(ckHex) != 40 {
		fatal(".cryptkey 解出的内容格式不对（邮箱密码多半不对）")
	}
	cryptKey, err := hex.DecodeString(ckHex)
	if err != nil {
		fatal(".cryptkey 不是十六进制串: %v", err)
	}
	fmt.Println("cryptkey 已解开")

	var accounts []mail.Account
	mainDS, err := loadDomain(*smDir, mainDomain)
	if err != nil {
		fatal("%v", err)
	}
	accounts = append(accounts, makeAccount(mainEmail, mainEmail, mailPass, mainDS, mainDomain))

	addlRaw, err := os.ReadFile(filepath.Join(mainDir, "additionalaccounts"))
	if err == nil {
		var addl map[string]addlAccount
		if err := json.Unmarshal(addlRaw, &addl); err != nil {
			fatal("解析 additionalaccounts: %v", err)
		}
		for _, a := range addl {
			domain := a.Email[strings.LastIndex(a.Email, "@")+1:]
			if a.HMAC != "" && !smcrypt.VerifyHMAC(cryptKey, a.Pass, a.HMAC) {
				fatal("账号 %s 的 hmac 校验不符，数据可能损坏", a.Email)
			}
			pv, err := smcrypt.DecryptUrlSafe(a.Pass, cryptKey, appSalt)
			if err != nil {
				fatal("解开账号 %s 的密码失败: %v", a.Email, err)
			}
			pass, ok := pv.(string)
			if !ok || pass == "" {
				fatal("账号 %s 解出的密码格式不对", a.Email)
			}
			ds, err := loadDomain(*smDir, domain)
			if err != nil {
				fatal("%v", err)
			}
			accounts = append(accounts, makeAccount(a.Email, a.Login, pass, ds, domain))
			fmt.Printf("附加账号 %s 密码已解开\n", a.Email)
		}
	}

	var verified []mail.Account
	for _, a := range accounts {
		cl, err := mail.Connect(a, nil)
		if err != nil {
			fmt.Printf("跳过 %s: IMAP 登录验证失败: %v\n", a.Email, err)
			continue
		}
		cl.Close()
		fmt.Printf("验证通过: %s (%s:%d)\n", a.Email, a.ImapHost, a.ImapPort)
		verified = append(verified, a)
	}
	if len(verified) == 0 {
		fatal("没有任何账号通过登录验证，未做任何更改")
	}

	st, err := store.Open(*dbPath)
	if err != nil {
		fatal("打开 webmail 数据库: %v", err)
	}
	defer st.DB.Close()

	existing := map[string]bool{}
	rows, err := st.DB.Query("SELECT email FROM accounts")
	if err != nil {
		fatal("查询已有账号: %v", err)
	}
	for rows.Next() {
		var e string
		if err := rows.Scan(&e); err == nil {
			existing[e] = true
		}
	}
	rows.Close()

	mainVerified := false
	var fresh []mail.Account
	for _, a := range verified {
		if a.Email == mainEmail {
			mainVerified = true
		}
		if !existing[a.Email] {
			fresh = append(fresh, a)
		}
	}
	if !existing[mainEmail] && !mainVerified {
		fatal("主账号 %s 既不在 webmail 里、这次也没通过登录验证，未做任何更改", mainEmail)
	}
	if len(fresh) == 0 {
		fmt.Println("没有新账号可导入：验证通过的账号都已存在。")
		return
	}

	var key []byte
	if *keepMaster {
		cur, err := readTTYPassword("请输入当前 webmail 主密码（输入不回显）: ")
		if err != nil {
			fatal("%v", err)
		}
		k, ok, err := auth.VerifyMaster(st, cur)
		if err != nil || !ok {
			fatal("当前主密码不对，未做任何更改")
		}
		key = k
	} else {
		newMaster, err := readTTYPassword("请设置新的 webmail 主密码（至少 8 位，输入不回显）: ")
		if err != nil {
			fatal("%v", err)
		}
		if len(newMaster) < 8 {
			fatal("新主密码太短，未做任何更改")
		}
		again, err := readTTYPassword("再输入一次新主密码: ")
		if err != nil {
			fatal("%v", err)
		}
		if again != newMaster {
			fatal("两次输入的新主密码不一致，未做任何更改")
		}
		if err := auth.SetupMaster(st, newMaster); err != nil {
			fatal("设置新主密码: %v", err)
		}
		k, ok, err := auth.VerifyMaster(st, newMaster)
		if err != nil || !ok {
			fatal("新主密码校验失败: %v", err)
		}
		key = k
		if _, err := st.DB.Exec("DELETE FROM sessions"); err != nil {
			fatal("清理旧会话: %v", err)
		}
	}

	inserted := 0
	for _, a := range fresh {
		enc, err := auth.Encrypt(key, a.Password)
		if err != nil {
			fatal("加密 %s 的密码: %v", a.Email, err)
		}
		_, err = st.DB.Exec(`INSERT INTO accounts
			(label, email, imap_host, imap_port, imap_tls, smtp_host, smtp_port, smtp_tls, username, enc_password, sort, created_at)
			VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?)`,
			a.Label, a.Email, a.ImapHost, a.ImapPort, a.ImapTLS,
			a.SMTPHost, a.SMTPPort, a.SMTPTLS, a.Username, enc, time.Now().Unix())
		if err != nil {
			fatal("写入账号 %s: %v", a.Email, err)
		}
		inserted++
		fmt.Printf("已导入: %s\n", a.Email)
	}
	fmt.Printf("完成: 共导入 %d 个账号。登录 webmail 后账号会自动开始同步。\n", inserted)
}
