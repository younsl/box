## Kubernetes 자동완성

```bash
brew install npm
```

`npm`을 설치합니다. `npm`이 없을 경우 `vim`을 실행할 떄마다 자동 업데이트 체크 과정에서 `[coc.nvim] Can't find npm or yarn in your $PATH"` 에러가 발생합니다.

```bash
:CocInstall coc-yaml
```

```bash
:CocConfig
```

`.yaml` 파일의 경우 Kubernetes 자동완성 추가

```bash
{
    "languageserver": {
        "terraform": {
            "command": "terraform-ls",
            "args": ["serve"],
            "filetypes": [
                "terraform",
                "tf"
            ],
            "initializationOptions": {},
            "settings": {}
        }
    },
    "yaml.schemas": {
        "kubernetes": "/*.yaml"
    }
}
```

---

## terraform 자동완성

자동완성 플러그인인 CoC(언어 서버)를 실행하려면 Node.js를 설치해야 합니다.

```bash
brew install npm
```

테라폼 공식 언어서버를 설치합니다.

```bash
brew install hashicorp/tap/terraform-ls
```

CocConfig에 언어 서버 구성 추가

```json
{
    "languageserver": {
        "terraform": {
            "command": "terraform-ls",
            "args": ["serve"],
            "filetypes": [
                "terraform",
                "tf"
            ],
            "initializationOptions": {},
            "settings": {}
        }
    }
}
```

---

## Syntax Highlighting

테라폼 코드의 syntax highlighting이 필요한 경우 vim-polyglot 플러그인을 설치해서 사용하면 됩니다.

vim-plug 설치


```bash
curl -fLo ~/.vim/autoload/plug.vim --create-dirs \
  https://raw.githubusercontent.com/junegunn/vim-plug/master/plug.vim
```

vim-plug는 vim 전용 플러그인 매니저입니다.

&nbsp;

`~/.vim/.vimrc`에 [vim-polyglot](https://github.com/sheerun/vim-polyglot) 플러그인 추가

```bash
set nocompatible

call plug#begin()

Plug 'sheerun/vim-polyglot'

call plug#end()
```

`vim`을 켠 후 플러그인을 설치합니다.

```bash
:PlugInstall
```

---

## 관련자료

[Setting up Vim for YAML editing](https://www.arthurkoziel.com/setting-up-vim-for-yaml/)
