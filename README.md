# wexe

一个监听文件变动或监听时间而触发预设动作的工具

## 使用示例

### 热编译

```bash
wexe --wf=*.go --cmd="go build"
```

### 定时同步文件

```bash
wexe --wt=10s --cmd="rsync -avz src dst"
```