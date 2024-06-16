curl -X POST http://localhost:8080/go_load.GoLoadService/CreateDownloadTask \
  -H "Content-Type: application/json" \
  -b "GOLOAD_AUTH=eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MTg1NTAxNDcsImtpZCI6MTMsInN1YiI6MX0.bi6r3yAqS0iTindAZh2bpLAdfYhOROE_rQef-ZJXnCHAQ-KPfQVlwuSI4-yJ2L8afqlAEUDo-NLikCdWBlndi_tG2iK5RknQ_im3RNfqNpTq7oXU7G_76oTwsi-RShbAG47XfDxznYCQvXm3LJpIjCNk8_IlZzwIJESoTtFI99HmOA8XPeLjcPaOjLqm2jk7B4HXXFnk1VH1OBrceCYW95nhUpyg9xKiuuDWsfRcQz2nNSIpLyXdTM0aCqD6JCHiZK1-m4V-jQeLjrmFIRSKXQQmQRl3VBqalfYRshvTnnV4r52OvjFuwg1FOZ2MT_3EbtEPkRYgCWDlBectf2BDdg" \
  -d '{"url": "example.com/download/abc", "download_type": 1}'
