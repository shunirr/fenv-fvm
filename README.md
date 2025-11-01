# fenv-fvm

fenv-fvmは、[fenv](https://github.com/fenv-org/fenv)互換のFlutterバージョン管理ツールです。バックエンドに[FVM (Flutter Version Management)](https://fvm.app/)を使用し、CI環境でfenvと同じワークフローを実現します。

## 特徴

- **単一バイナリ**: DartやFlutterのランタイム不要
- **fenv互換**: `.flutter-version`ファイルでバージョン管理
- **FVMバックエンド**: FVMの強力なSDK管理機能を活用
- **CI/CD最適**: Codemagicなどfvm対応CI環境で即座に使用可能
- **マルチプラットフォーム**: Linux (x86_64/aarch64) および macOS (x86_64/arm64) 対応

## 必要要件

- `fvm`がPATHに存在すること
- ネットワーク接続（Flutter SDKダウンロード用）

## インストール

### GitHubリリースからダウンロード

最新リリースから、お使いのプラットフォームに対応したバイナリをダウンロードしてください：

```bash
# 例: macOS arm64の場合
curl -L -o fenv-fvm.tar.gz https://github.com/YOUR_USERNAME/fenv-fvm/releases/latest/download/fenv-fvm-darwin-arm64.tar.gz
tar -xzf fenv-fvm.tar.gz
chmod +x fenv-fvm
sudo mv fenv-fvm /usr/local/bin/
```

利用可能なバイナリ：
- `fenv-fvm-linux-amd64.tar.gz`
- `fenv-fvm-linux-aarch64.tar.gz`
- `fenv-fvm-darwin-amd64.tar.gz` (Intel Mac)
- `fenv-fvm-darwin-arm64.tar.gz` (Apple Silicon)

## セットアップ

### 1. Shimの初期化

fenv-fvmのshimをセットアップし、PATHに追加します：

```bash
eval "$(fenv-fvm init)"
```

シェルの設定ファイル（`~/.bashrc`, `~/.zshrc`など）に追加することで、永続的に有効化できます：

```bash
echo 'eval "$(fenv-fvm init)"' >> ~/.zshrc
```

### 2. プロジェクトでFlutterバージョンを指定

プロジェクトルートに`.flutter-version`ファイルを作成します：

```bash
# 方法1: fenv-fvm localコマンドで作成
fenv-fvm local 3.13.9

# 方法2: 手動で作成
echo "3.13.9" > .flutter-version
```

### 3. Flutterコマンドをそのまま使用

以降、通常の`flutter`や`dart`コマンドが自動的に正しいバージョンで実行されます：

```bash
flutter --version
flutter pub get
flutter build apk
dart --version
```

## コマンドリファレンス

### `fenv-fvm init`

shimディレクトリをセットアップし、PATH設定用のシェルスクリプトを出力します。

```bash
eval "$(fenv-fvm init)"
```

### `fenv-fvm local [version]`

#### バージョン指定あり

現在のディレクトリに`.flutter-version`ファイルを作成し、指定したFlutterバージョンをインストールします。

```bash
fenv-fvm local 3.13.9
fenv-fvm local stable
```

#### バージョン指定なし

既存の`.flutter-version`ファイルを読み取り、FVMでSDKを同期します（主にCI用）。

```bash
fenv-fvm local
```

### `fenv-fvm install <version>`

指定したFlutterバージョンを事前にダウンロードします。`.flutter-version`ファイルは変更しません。

```bash
fenv-fvm install 3.13.9
```

### `fenv-fvm version`

現在のプロジェクトで設定されているFlutterバージョンを表示します。

```bash
fenv-fvm version
# 出力例: 3.13.9 (set by /path/to/project/.flutter-version)
```

## CI/CD環境での使用

### 典型的なCI設定例

```yaml
# .github/workflows/build.yml の例
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install FVM
        run: |
          dart pub global activate fvm
          echo "$HOME/.pub-cache/bin" >> $GITHUB_PATH

      - name: Install fenv-fvm
        run: |
          curl -L -o fenv-fvm.tar.gz https://github.com/YOUR_USERNAME/fenv-fvm/releases/latest/download/fenv-fvm-linux-amd64.tar.gz
          tar -xzf fenv-fvm.tar.gz
          chmod +x fenv-fvm
          sudo mv fenv-fvm /usr/local/bin/

      - name: Setup Flutter
        run: |
          eval "$(fenv-fvm init)"
          fenv-fvm local

      - name: Build
        run: |
          eval "$(fenv-fvm init)"
          flutter pub get
          flutter build apk
```

### Codemagic

Codemagicではfvmがプリインストールされているため、さらにシンプルです：

```yaml
workflows:
  build:
    environment:
      flutter: stable
    scripts:
      - name: Setup fenv-fvm
        script: |
          curl -L -o fenv-fvm.tar.gz https://github.com/YOUR_USERNAME/fenv-fvm/releases/latest/download/fenv-fvm-linux-amd64.tar.gz
          tar -xzf fenv-fvm.tar.gz
          chmod +x fenv-fvm
          export PATH="$PWD:$PATH"
          eval "$(fenv-fvm init)"
          fenv-fvm local

      - name: Build
        script: |
          eval "$(fenv-fvm init)"
          flutter build apk
```

## 仕組み

fenv-fvmは以下のように動作します：

1. `.flutter-version`ファイルからFlutterバージョンを読み取る
2. `fvm install <version>`と`fvm use <version>`を実行してSDKを準備
3. `<project>/.fvm/flutter_sdk/bin/flutter`（または`dart`）へのパスを解決
4. 現在のプロセスを解決されたバイナリで置き換え（`syscall.Exec`）

これにより、通常の`flutter`/`dart`コマンドが透過的に適切なバージョンで実行されます。

## トラブルシューティング

### `fvm not found in PATH`

fvmがインストールされていないか、PATHに含まれていません：

```bash
# fvmのインストール
dart pub global activate fvm

# PATHに追加
export PATH="$PATH:$HOME/.pub-cache/bin"
```

### `.flutter-version not found`

プロジェクトルートに`.flutter-version`ファイルが存在しません：

```bash
fenv-fvm local 3.13.9
```

## ライセンス

MIT License

## 関連プロジェクト

- [fenv](https://github.com/fenv-org/fenv) - オリジナルのfenv
- [FVM](https://fvm.app/) - Flutter Version Management