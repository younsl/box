nvidia-device-plugin 데몬셋보다는 gpu-operator를 사용하면 운영이 편함

gpu-operator 차트는 git clone 하지말고 helm pull로 저장소에서 직접 다운로드 받아야 정상 동작함

```bash
helm pull nvidia/gpu-operator --version 25.3.1 --untar
```
