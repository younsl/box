# Serverless Reverse Proxy

EC2 서버 없이 서버리스로 리버스 프록시를 구현할 수 있는 방법을 찾아보자.

결과적으로 Public ALB를 Internal NLB에 연결하는 방식으로 구현할 수 있음

![Architecture of Serverless Reverse Proxy](./assets/reverse-proxy-1.png)

자세한 내용은 아래 링크를 참고

- [Serverless Ingress Solution on AWS](https://jackiechen.blog/2022/09/07/serverless-ingress-solution-on-aws/)