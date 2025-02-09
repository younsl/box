#!/bin/bash

# 도움말 출력 함수
print_usage() {
    echo "사용법: $0 [--dry-run]"
    echo "  --dry-run    실제 PDF 생성 없이 실행될 명령어만 출력"
}

# 인자 파싱
DRY_RUN=0
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --dry-run) DRY_RUN=1 ;;
        -h|--help) print_usage; exit 0 ;;
        *) echo "알 수 없는 옵션: $1"; print_usage; exit 1 ;;
    esac
    shift
done

# 현재 디렉토리에서 모든 하위 디렉토리를 순회
find . -type f -name "slide-deck.md" | while read slide_path; do
    dir_path=$(dirname "$slide_path")
    
    if [ $DRY_RUN -eq 1 ]; then
        echo "[DRY-RUN] 다음 명령어가 실행됩니다:"
        echo "cd $dir_path && docker run --rm --init -v $PWD:/home/marp/app/ -e LANG=$LANG marpteam/marp-cli slide-deck.md --pdf --allow-local-files"
    else
        echo "Converting: $slide_path"
        cd "$dir_path" && \
        podman run --rm --init \
            -v $PWD:/home/marp/app/ \
            -e LANG=$LANG \
            marpteam/marp-cli \
            slide-deck.md --pdf --allow-local-files
        cd - > /dev/null
    fi
done