# go-notice-temperature-discord

気温を取得して Discord に通知するプログラム
Raspberry Pi で動作させることを想定

## 環境変数

`NOTICE_TEMPERATURE_DISCORD_URL`: 通知先の Discord Webhook URL

## セットアップ

```
# ビルド
$ GOOS=linux GOARCH=arm go build -o notice-temperature-discord

# 実行ファイルを Raspberry Pi に転送
$ scp notice-temperature-discord pi5:/home/YOUR_NAME/
```

## つまづき

Cron で実行させようとした際に、環境変数 `NOTICE_TEMPERATURE_DISCORD_URL` が読み込まれていないという問題が発生した
結果としては、 `/home/YOUR_NAME` 配下に `discord.env` を作成して環境変数を記載し、Cron で実行する際に `.` コマンドで読み込むようにすることで解決した

```diff
-* * * * * /home/YOUR_NAME/notice-temperature-discord
+* * * * * . /home/YOUR_NAME/discord.env && /home/YOUR_NAME/notice-temperature-discord
```
