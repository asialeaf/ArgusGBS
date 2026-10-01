# livecms-web-src

LiveGBS 前端工程源码

# 起步

```
# 准备工具 nvm 切换 node 版本到 10.24.1
# nvm for linux => https://github.com/nvm-sh/nvm/blob/master/README.md
# nvm for linux => NVM_NODEJS_ORG_MIRROR=https://npmmirror.com/mirrors/node nvm install 10.24.1
# nvm for windows => https://github.com/coreybutler/nvm-windows
# nvm for windows => nvm node_mirror https://npmmirror.com/mirrors/node
# nvm for windows => nvm install 10.24.1
# nvm use 10.24.1

npm install -g cross-env

cd web
# 安装 npm 依赖
npm install
# 或
npm install --registry=https://registry.npmmirror.com --sass_binary_site=https://npmmirror.com/mirrors/node-sass

# 编辑 webpack.config.js > devServer > proxy target ip 指向 server ip

# 启动
npm run start
# 编译
npm run build
```

# 调主题色

1. web > assets > styles > element-custom.scss > $--color-primary

2. web > assets > styles > variables.less > @base