package main

import (
	"context"
	"crypto/tls"
	"embed"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	ipfs "github.com/ipfs/go-ipfs-api"
	"github.com/jeandeaual/go-locale"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed translations/*.toml
var translations embed.FS

func printOptError(localizer *i18n.Localizer, err error) {
	_, _ = fmt.Fprintln(os.Stderr, localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "Messages.IgnoredError",
		TemplateData: map[string]string{
			"Err": err.Error(),
		},
	}))
}

func main() {
	userLocales, err := locale.GetLocales()
	if err != nil {
		panic(err)
	}
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	_, _ = bundle.LoadMessageFileFS(translations, "translations/en.toml")
	_, _ = bundle.LoadMessageFileFS(translations, "translations/bn.toml")
	localizer := i18n.NewLocalizer(bundle, userLocales...)

	var certFile, certKey, shellURL, ipnsKey, dirToUpload string
	flag.StringVar(&certFile, "cert-file", "cert.crt", localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "Arguments.CertificatePath",
	}))
	flag.StringVar(&certKey, "private-key", "cert.key", localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "Arguments.PrivateKey",
	}))
	flag.StringVar(&shellURL, "shell-url", "https://localhost", localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "Arguments.ShellURL",
	}))
	flag.StringVar(&ipnsKey, "ipns-key", "", localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "Arguments.IPNSKey",
	}))
	flag.StringVar(&dirToUpload, "target", "dist", localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "Arguments.Target",
	}))
	flag.Parse()

	cert, err := tls.LoadX509KeyPair(certFile, certKey)
	if err != nil {
		panic(err)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		},
	}
	shell := ipfs.NewShellWithClient(shellURL, httpClient)

	id, err := shell.ID()
	if err != nil {
		panic(err)
	}
	_, _ = fmt.Fprintln(os.Stderr, localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "Messages.PublishMsg",
		TemplateData: map[string]string{
			"ID": id.ID,
		},
	}))

	var keyId string
	{
		keyList, err := shell.KeyList(context.TODO())
		if err != nil {
			panic(err)
		}
		for _, k := range keyList {
			if k.Name == ipnsKey {
				keyId = k.Id
			}
		}
		if len(keyId) == 0 {
			panic(errors.New(localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: "Messages.CouldntFindKeyID",
			})))
		}
	}

	// We have to clean up previous publishes
	oldCID, err := shell.Resolve(keyId)
	if err != nil {
		printOptError(localizer, err)
		goto cleanedUp
	}
	_, _ = fmt.Fprintln(os.Stderr, localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "Messages.OldCID",
		TemplateData: map[string]string{
			"CID": oldCID,
		},
	}))
	err = shell.Unpin(oldCID)
	if err != nil {
		printOptError(localizer, err)
		goto cleanedUp
	}
	err = shell.FilesRm(context.TODO(), oldCID, true)
	if err != nil {
		printOptError(localizer, err)
		goto cleanedUp
	}

cleanedUp:
	rootCID, err := shell.AddDir(dirToUpload)
	if err != nil {
		panic(err)
	}
	err = shell.Pin(rootCID)
	if err != nil {
		panic(err)
	}
	_, err = shell.PublishWithDetails(rootCID, keyId, 24*time.Hour, 24*time.Hour, true)
	if err != nil {
		panic(err)
	}
	_, _ = fmt.Fprintln(os.Stderr, localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "Messages.Published",
		TemplateData: map[string]string{
			"CID": rootCID,
			"Key": keyId,
		},
	}))
}
