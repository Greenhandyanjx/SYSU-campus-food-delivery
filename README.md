# SYSU-campus-food-delivery
中珠外卖校园送优化系统
开发流程（每个人日常操作）
克隆项目（只做一次）
git clone git@github.com:xxx/SYSU-campus-food-delivery
.git
cd SYSU-campus-food-delivery
拉最新开发分支
git checkout dev
git pull origin dev
创建自己的功能分支
git checkout -b feature/frontend-homepage
写代码 + 提交
git add .
git commit -m "feat: add homepage layout"
推送到远程
git push origin feature/frontend-homepage
发起 Pll Request (PR)
在 GitHub 网页上点 New Pull Request
目标分支：dev
由组长 review → merge
开发前一定先同步最新代码（避免冲突
git checkout dev
git pull origin dev
git checkout feature/frontend-homepage
git merge dev   # 把最新dev合并到你的分支
