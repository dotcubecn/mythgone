<div align="center">
  <a href="#"><img src="docs/mythgone256.png" alt="Mythgone 图标" width="128" height="128"></a>
  <br><sub>图标基于 <a href="https://egonelbre.com/">Egon Elbre</a> 的 Go gopher 形象创作, 其中引用的极域电子教室图标, 其版权与商标权归南京极域信息科技有限公司所有.</sub>
  <h3 align="center">Mythgone</h3>
  <p>适用于 Windows 的简洁极域电子教室反控软件, 使用 Go 编写.</p>
  <a href="#"><img alt="徽章: Go 语言版本" src="https://img.shields.io/github/go-mod/go-version/dotcubecn/mythgone?style=square"></a>
  <a href="https://github.com/dotcubecn/mythgone?tab=GPL-3.0-1-ov-file"><img alt="徽章: GPL 3.0 许可证" src="https://img.shields.io/github/license/dotcubecn/mythgone?style=square"></a>
  <a href="https://github.com/dotcubecn/mythgone/commits"><img alt="徽章: 最后提交时间" src="https://img.shields.io/github/last-commit/dotcubecn/mythgone?style=square"></a>
  <a href="https://github.com/dotcubecn/mythgone/issues"><img alt="徽章: 议题数" src="https://img.shields.io/github/issues/dotcubecn/mythgone?style=square"></a>
  <a href="https://github.com/dotcubecn/mythgone/stargazers"><img alt="徽章: 星标数" src="https://img.shields.io/github/stars/dotcubecn/mythgone?style=square"></a>
  <a href="#"><img alt="徽章: 项目大小" src="https://img.shields.io/github/repo-size/dotcubecn/mythgone?style=square&label=size"></a>
</div>

---

## 兼容性
本软件在如下环境中经测试可正常运行.  
| **操作系统** | **架构** | **极域版本** | **备注** |
| :--- | :--- | :--- | :--- |
| Windows 7 旗舰版 SP1 | x86 | v4.2 2015 专业版 | 部分功能受限 |
| Windows 10 IoT 企业版 LTSC 2021 | x64 | v6.0 2016 豪华版 |  |
| Windows 10 IoT 企业版 LTSC 2021 | x64 | v6.0 2021 豪华版 |  |

## 文档
> [!WARNING]  
> 文档内容由 AI 生成, 可能存在描述不准确或更新延迟的情况, 请以项目实际代码为准.

[![zread](https://img.shields.io/badge/Ask_Zread-_.svg?style=flat&color=00b0aa&labelColor=000000&logo=data%3Aimage%2Fsvg%2Bxml%3Bbase64%2CPHN2ZyB3aWR0aD0iMTYiIGhlaWdodD0iMTYiIHZpZXdCb3g9IjAgMCAxNiAxNiIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTQuOTYxNTYgMS42MDAxSDIuMjQxNTZDMS44ODgxIDEuNjAwMSAxLjYwMTU2IDEuODg2NjQgMS42MDE1NiAyLjI0MDFWNC45NjAxQzEuNjAxNTYgNS4zMTM1NiAxLjg4ODEgNS42MDAxIDIuMjQxNTYgNS42MDAxSDQuOTYxNTZDNS4zMTUwMiA1LjYwMDEgNS42MDE1NiA1LjMxMzU2IDUuNjAxNTYgNC45NjAxVjIuMjQwMUM1LjYwMTU2IDEuODg2NjQgNS4zMTUwMiAxLjYwMDEgNC45NjE1NiAxLjYwMDFaIiBmaWxsPSIjZmZmIi8%2BCjxwYXRoIGQ9Ik00Ljk2MTU2IDEwLjM5OTlIMi4yNDE1NkMxLjg4ODEgMTAuMzk5OSAxLjYwMTU2IDEwLjY4NjQgMS42MDE1NiAxMS4wMzk5VjEzLjc1OTlDMS42MDE1NiAxNC4xMTM0IDEuODg4MSAxNC4zOTk5IDIuMjQxNTYgMTQuMzk5OUg0Ljk2MTU2QzUuMzE1MDIgMTQuMzk5OSA1LjYwMTU2IDE0LjExMzQgNS42MDE1NiAxMy43NTk5VjExLjAzOTlDNS42MDE1NiAxMC42ODY0IDUuMzE1MDIgMTAuMzk5OSA0Ljk2MTU2IDEwLjM5OTlaIiBmaWxsPSIjZmZmIi8%2BCjxwYXRoIGQ9Ik0xMy43NTg0IDEuNjAwMUgxMS4wMzg0QzEwLjY4NSAxLjYwMDEgMTAuMzk4NCAxLjg4NjY0IDEwLjM5ODQgMi4yNDAxVjQuOTYwMUMxMC4zOTg0IDUuMzEzNTYgMTAuNjg1IDUuNjAwMSAxMS4wMzg0IDUuNjAwMUgxMy43NTg0QzE0LjExMTkgNS42MDAxIDE0LjM5ODQgNS4zMTM1NiAxNC4zOTg0IDQuOTYwMVYyLjI0MDFDMTQuMzk4NCAxLjg4NjY0IDE0LjExMTkgMS42MDAxIDEzLjc1ODQgMS42MDAxWiIgZmlsbD0iI2ZmZiIvPgo8cGF0aCBkPSJNNCAxMkwxMiA0TDQgMTJaIiBmaWxsPSIjZmZmIi8%2BCjxwYXRoIGQ9Ik00IDEyTDEyIDQiIHN0cm9rZT0iI2ZmZiIgc3Ryb2tlLXdpZHRoPSIxLjUiIHN0cm9rZS1saW5lY2FwPSJyb3VuZCIvPgo8L3N2Zz4K&logoColor=ffffff)](https://zread.ai/dotcubecn/mythgone)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/dotcubecn/mythgone)
- [中文文档 (Zread)](https://zread.ai/dotcubecn/mythgone)
- [英文文档 (DeepWiki)](https://deepwiki.com/dotcubecn/mythgone)

## 致谢
### 鸣谢
- [@数码小-a](https://space.bilibili.com/3546704078964833 "数码小-a的哔哩哔哩主页") - 提供了用于测试的极域样本.

### 第三方库
- [thongtech/go-legacy-win7](https://github.com/thongtech/go-legacy-win7 "go-legacy-win7 的 GitHub 仓库") - 使 Go 语言支持旧版本 Windows.  
- [tailscale/walk](https://github.com/tailscale/walk "tailscale 在 GitHub 上的 walk 分支 ") - 用于创建 Windows GUI.  

### 参考
- [极域电子教室完全分 (附绕过方法和软件)](https://www.52pojie.cn/thread-542884-1-1.html "极域电子教室完全分析（附绕过方法和软件） - 吾爱破解 - 52pojie.cn")
- [对极域64位禁止终止进程、键盘锁定的分析](https://blog.csdn.net/weixin_42112038/article/details/126228989 "对极域64位禁止终止进程、键盘锁定的分析_请求的控件对此服务无效-CSDN博客")
- [Windows API 让窗口在截屏 (极域监控) 时变透明 窗口防截屏 让老师祖传十年的极域瞎掉](https://www.cnblogs.com/petyr/articles/19001342 "Windows API 让窗口在截屏（极域监控）时变透明 窗口防截屏 让老师祖传十年的极域瞎掉 - Petyrma - 博客园")
- [win32 判断进程状态 (挂起/运行中), 用API挂起/恢复进程](https://blog.csdn.net/weixin_42112038/article/details/126243863 "win32 判断进程状态（挂起/运行中）、用API挂起/恢复进程_判断一个进程是否处于挂起状态-CSDN博客")
- [极域电子教室 v6.0 部分 UDP 数据包分析](https://gist.github.com/dotcubecn/57bdd9578f105f20009cc6fd2b64f4da "极域 v6.0 部分 UDP 数据包分析")

## 赞助
如果本软件对你有帮助, 你可以考虑赞助我.  
> [!CAUTION]  
> **请在个人经济条件允许的情况下赞助, 未成年人请在监护人同意后赞助. 请勿支出超出自身承受能力的费用.**

[![爱发电赞助](https://img.shields.io/badge/爱发电-赞助开发者-946ce6?style=for-the-badge)](https://ifdian.net/order/create?user_id=1c339020ef8111ec9f4752540025c377)
[![硅基流动推广](https://img.shields.io/badge/硅基流动-免费额度-9354ff?style=for-the-badge)](https://cloud.siliconflow.cn/i/qqpGopO3)
[![雨云推广](https://img.shields.io/badge/雨云-优惠注册-37b5c1?style=for-the-badge)](https://www.rainyun.com/dotcube_?s=gh-mythgone-readme)

![赞助二维码](docs/sponsor.webp "赞助二维码")
