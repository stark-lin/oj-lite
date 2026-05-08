# Lua 基础算法双语课程教学文档

## 文档说明

本文档用于规划一套面向 Scratch 进阶学生的 Lua 基础算法课程。课程采用中英文双语 lesson 配套设计，同一周的英文 lesson 使用奇数编号，中文 lesson 使用偶数编号。

每个 lesson 的序号固定，用于排序和引用；lesson 名称只使用“周数 + 描述”，不把语言和题目难度写进 lesson 名称。

编号规则：

```text
Lesson 01, 03, 05, ... 23：English lesson
Lesson 02, 04, 06, ... 24：中文 lesson
```

## 1. 课程定位

本课程面向已经学完 Scratch、准备过渡到文本编程的学生。课程用 Lua 作为表达工具，重点训练基础算法思维，为后续学习 Python / C / C++ 打基础。

课程原则：

| 原则 | 说明 |
| --- | --- |
| 算法优先 | Lua 只承担表达算法的作用，不把课程做成 Lua 语法课 |
| 简洁递进 | 每节课只引入少量新概念，题目保持短小、明确、可验证 |
| 场景化 | 所有题目来自学生熟悉的游戏、校园、购物、运动、成绩等场景 |
| 双语配套 | 每周一套中文 lesson，一套英文 lesson，结构和知识点保持一致 |
| 固定结构 | 每套 lesson 分为 Simple、Normal、Challenge 三档 |

## 2. 技术边界

不使用：

```lua
io.read
string.*
table.insert
table.sort
pairs
ipairs
math.*
```

只使用：

```lua
local 变量
if / elseif / else
for
while
function
数组 table
数组下标
基本四则运算
比较运算
逻辑运算
```

## 3. 总体结构

课程共 12 周，每周 2 套 lesson。同一主题先安排 English lesson，再安排中文 lesson：

```text
English lesson
中文 lesson
```

总计：

```text
12 周 x 2 套语言 lesson = 24 套 lesson
```

每套 lesson 固定 9 题：

| 难度 | 数量 | 作用 |
| --- | ---: | --- |
| Simple | 3 题 | 热身，确认概念 |
| Normal | 3 题 | 主线训练，完成本周核心算法 |
| Challenge | 3 题 | 拔高、综合应用或课后作业 |

总题量：

```text
每周中文 9 题 + 英文 9 题 = 18 题
12 周共 108 题中文 + 108 题英文 = 216 题
```

## 4. 24 套 Lesson 编号顺序

| Lesson 序号 | 语言 | Lesson 名称 |
| ---: | --- | --- |
| 01 | EN | Week 01 - Variables |
| 02 | CN | Week 01 - 变量 |
| 03 | EN | Week 02 - Conditions and Loops |
| 04 | CN | Week 02 - 判断与循环 |
| 05 | EN | Week 03 - Counting and Classification |
| 06 | CN | Week 03 - 计数与分类统计 |
| 07 | EN | Week 04 - Sum, Average, and Accumulation |
| 08 | CN | Week 04 - 累加、平均与累计 |
| 09 | EN | Week 05 - Factors, Multiples, and Prime Numbers |
| 10 | CN | Week 05 - 因数、倍数与质数 |
| 11 | EN | Week 06 - Quantity Relations and Double Enumeration |
| 12 | CN | Week 06 - 数量关系与双变量枚举 |
| 13 | EN | Week 07 - Arrays and Table Data |
| 14 | CN | Week 07 - 数组与表格数据 |
| 15 | EN | Week 08 - Array Statistics and Extremes |
| 16 | CN | Week 08 - 数组统计与最值 |
| 17 | EN | Week 09 - Search and Binary Search |
| 18 | CN | Week 09 - 查找与二分查找 |
| 19 | EN | Week 10 - Sorting Algorithms I |
| 20 | CN | Week 10 - 排序算法一 |
| 21 | EN | Week 11 - Sorting Algorithms II and Applications |
| 22 | CN | Week 11 - 排序算法二与排序应用 |
| 23 | EN | Week 12 - Integrated Math Algorithm Projects |
| 24 | CN | Week 12 - 综合数学算法项目 |

## 5. 12 周课程主线

| 周次 | 主题 | 核心算法能力 |
| ---: | --- | --- |
| Week 01 | 变量 | 顺序计算、变量更新、变量交换 |
| Week 02 | 判断与循环 | 条件分类、循环计数、边界判断 |
| Week 03 | 计数与分类统计 | 条件计数、分类统计 |
| Week 04 | 累加、平均与累计 | 求和、条件求和、平均值、累计 |
| Week 05 | 因数、倍数与质数 | 枚举、因数计数、质数判断 |
| Week 06 | 数量关系与双变量枚举 | 双重循环、方案搜索、组合计数 |
| Week 07 | 数组与表格数据 | 数组访问、修改、遍历、复制 |
| Week 08 | 数组统计与最值 | 数组求和、计数、最大值、最小值、第二大值 |
| Week 09 | 查找与二分查找 | 线性查找、位置查找、二分查找 |
| Week 10 | 排序算法一 | 选择排序、冒泡排序 |
| Week 11 | 排序算法二与排序应用 | 插入排序、排序后处理、合并有序数组 |
| Week 12 | 综合数学算法项目 | 统计、查找、排序、枚举、建模 |

## 6. 每周详细大纲

### Week 01：变量

| 项目 | 内容 |
| --- | --- |
| 核心主题 | 变量是保存数据的盒子 |
| 数学基础 | 四则运算、面积、周长、总价 |
| 编程内容 | `local`、赋值、变量更新、临时变量 |
| 算法内容 | 顺序计算、变量交换、状态更新 |
| 场景 | 游戏血量、操场面积、文具购物、金币变化、换座位 |
| 目标 | 学生能用变量表达一个具体应用题的计算过程 |

教学任务设计：

| 序号 | 难度 | Title |
| ---: | --- | --- |
| 01 | Simple | 游戏角色受到攻击后还剩多少血量 |
| 02 | Simple | 文具店买铅笔一共要花多少钱 |
| 03 | Simple | 操场长方形区域的面积是多少 |
| 04 | Normal | 打怪获得金币后背包里共有多少金币 |
| 05 | Normal | 买铅笔和橡皮后还剩多少钱 |
| 06 | Normal | 跑步训练后累计跑了多少米 |
| 07 | Challenge | 两个同学换座位后座位编号如何交换 |
| 08 | Challenge | 游戏角色先受伤再回血后剩多少血量 |
| 09 | Challenge | 商店打包购买后计算总价和找零 |

核心例子：

```lua
local hp = 100
local damage = 25
local remainHp = hp - damage
```

### Week 02：判断与循环

| 项目 | 内容 |
| --- | --- |
| 核心主题 | 条件决定分支，循环处理重复 |
| 数学基础 | 比较大小、奇偶、倍数、边界值 |
| 编程内容 | `if`、`else`、`elseif`、`for`、`while` |
| 算法内容 | 条件分类、循环计数、边界判断 |
| 场景 | 温度判断、学生编号、考试及格、怪物扣血、闰年判断 |
| 目标 | 学生能写出“如果……”和“重复……”两类基础算法结构 |

教学任务设计：

| 序号 | 难度 | Title |
| ---: | --- | --- |
| 01 | Simple | 判断考试成绩是否达到及格线 |
| 02 | Simple | 判断学生编号是奇数还是偶数 |
| 03 | Simple | 判断今天温度是否适合户外运动 |
| 04 | Normal | 怪物每回合扣血，几回合后被打败 |
| 05 | Normal | 统计 1 到 n 中有多少个偶数 |
| 06 | Normal | 判断一个数字是否是 3 的倍数 |
| 07 | Challenge | 根据年份判断学校运动会是否遇到闰年 |
| 08 | Challenge | 自动售货机根据金额判断能买哪种饮料 |
| 09 | Challenge | 游戏关卡根据分数判断获得几颗星 |

核心例子：

```lua
local score = 75
local passLine = 60
local passed = false

if score >= passLine then
    passed = true
else
    passed = false
end
```

### Week 03：计数与分类统计

| 项目 | 内容 |
| --- | --- |
| 核心主题 | 有多少个符合条件 |
| 数学基础 | 数数、分类、倍数、余数 |
| 编程内容 | 循环 + 条件 + `count` |
| 算法内容 | 计数算法、条件计数、分类统计 |
| 场景 | 座位编号、学生编号、任务编号、图书分级、抽奖号码 |
| 目标 | 学生掌握 `count = count + 1` 的使用场景 |

教学任务设计：

| 序号 | 难度 | Title |
| ---: | --- | --- |
| 01 | Simple | 统计班级里有多少个偶数座位号 |
| 02 | Simple | 统计 1 到 50 中有多少个 5 的倍数 |
| 03 | Simple | 统计任务编号中有多少个奇数任务 |
| 04 | Normal | 抽奖号码中有多少个幸运号码 |
| 05 | Normal | 图书馆按编号统计普通图书数量 |
| 06 | Normal | 统计运动员编号中能被 3 整除的人数 |
| 07 | Challenge | 图书馆按编号分类统计三种等级图书数量 |
| 08 | Challenge | 游戏任务按难度编号统计不同类型任务 |
| 09 | Challenge | 学生编号按余数分成三组并统计人数 |

核心例子：

```lua
local n = 100
local k = 3
local count = 0

for i = 1, n do
    if i % k == 0 then
        count = count + 1
    end
end
```

### Week 04：累加、平均与累计

| 项目 | 内容 |
| --- | --- |
| 核心主题 | 总共是多少，平均是多少，累计到哪里 |
| 数学基础 | 连加、平均数、累计数量 |
| 编程内容 | `sum`、循环累加、数组累计雏形 |
| 算法内容 | 求和、条件求和、平均值、累计表、前缀和雏形 |
| 场景 | 楼梯编号、关卡奖励、每日金币、存钱目标 |
| 目标 | 学生能区分 `count` 和 `sum` 的用途 |

教学任务设计：

| 序号 | 难度 | Title |
| ---: | --- | --- |
| 01 | Simple | 计算一周每天金币奖励的总数 |
| 02 | Simple | 计算 1 到 n 的楼梯编号总和 |
| 03 | Simple | 计算几天训练一共跑了多少米 |
| 04 | Normal | 计算班级几次小测的平均分 |
| 05 | Normal | 每天存钱，计算第 n 天一共存了多少 |
| 06 | Normal | 统计关卡奖励金币的总数和平均数 |
| 07 | Challenge | 每天存钱，最早第几天达到买玩具目标 |
| 08 | Challenge | 连续爬楼梯，到第几层累计台阶超过目标 |
| 09 | Challenge | 游戏每日奖励累计到哪一天超过升级要求 |

核心例子：

```lua
local n = 100
local sum = 0

for i = 1, n do
    sum = sum + i
end
```

### Week 05：因数、倍数与质数

| 项目 | 内容 |
| --- | --- |
| 核心主题 | 用枚举解决数的性质问题 |
| 数学基础 | 因数、倍数、质数、公因数、公倍数 |
| 编程内容 | `%`、循环枚举、布尔变量 |
| 算法内容 | 因数枚举、因数计数、质数判断、最大公因数枚举、最小公倍数枚举 |
| 场景 | 分组比赛、拼图分块、特殊编号、跑步集合 |
| 目标 | 学生理解“一个个试”的算法思想 |

教学任务设计：

| 序号 | 难度 | Title |
| ---: | --- | --- |
| 01 | Simple | 判断一个队伍人数能不能平均分成若干组 |
| 02 | Simple | 判断一个编号是否是另一个编号的倍数 |
| 03 | Simple | 找出一个拼图块数能否平均分给小组 |
| 04 | Normal | 找出一个拼图块数的所有分法 |
| 05 | Normal | 统计一个特殊编号有多少个因数 |
| 06 | Normal | 找出两个跑步周期第一次同时集合的时间 |
| 07 | Challenge | 判断一个特殊编号是不是质数编号 |
| 08 | Challenge | 找出两个班人数的最大公共分组人数 |
| 09 | Challenge | 找出两个闹钟第一次同时响起的分钟数 |

核心例子：

```lua
local n = 17
local factorCount = 0

for i = 1, n do
    if n % i == 0 then
        factorCount = factorCount + 1
    end
end

local isPrime = false

if factorCount == 2 then
    isPrime = true
end
```

### Week 06：数量关系与双变量枚举

| 项目 | 内容 |
| --- | --- |
| 核心主题 | 两个变量一起变化时如何找答案 |
| 数学基础 | 和差关系、乘法关系、鸡兔同笼、购物方案 |
| 编程内容 | 双重循环、条件筛选 |
| 算法内容 | 双变量枚举、组合计数、方案搜索 |
| 场景 | 分糖、长方形拼图、骰子游戏、文具购买、鸡兔同笼 |
| 目标 | 学生能用双重循环解决简单应用题 |

教学任务设计：

| 序号 | 难度 | Title |
| ---: | --- | --- |
| 01 | Simple | 两种糖果一共买 20 颗有多少种买法 |
| 02 | Simple | 两个骰子的点数和等于目标有多少种情况 |
| 03 | Simple | 长方形拼图周长固定时有多少种长宽组合 |
| 04 | Normal | 文具店用固定预算买铅笔和橡皮的方案数 |
| 05 | Normal | 游戏商店买药水和盾牌刚好花完金币的方案数 |
| 06 | Normal | 两个班合报名额固定时有多少种分配方案 |
| 07 | Challenge | 鸡兔同笼：根据头数和脚数找动物数量 |
| 08 | Challenge | 游乐园买儿童票和成人票刚好用完预算 |
| 09 | Challenge | 找出购买数量最多且不超过预算的购物方案 |

核心例子：

```lua
local target = 20
local count = 0

for a = 1, target do
    for b = 1, target do
        if a + b == target then
            count = count + 1
        end
    end
end
```

### Week 07：数组与表格数据

| 项目 | 内容 |
| --- | --- |
| 核心主题 | 一组数据如何存储和处理 |
| 数学基础 | 表格数据、顺序编号、位置 |
| 编程内容 | 数组 table、下标、遍历、修改 |
| 算法内容 | 数组访问、数组修改、数组遍历、数组复制、累计数组 |
| 场景 | 成绩单、体测记录、每日积分、队伍调头、传球游戏 |
| 目标 | 学生理解数组是“一排有编号的盒子” |

教学任务设计：

| 序号 | 难度 | Title |
| ---: | --- | --- |
| 01 | Simple | 读取成绩单中第几个学生的分数 |
| 02 | Simple | 读取体测表中第几个同学的跳远成绩 |
| 03 | Simple | 查看游戏背包中第几个格子的物品数量 |
| 04 | Normal | 修改每日积分表中某一天的积分 |
| 05 | Normal | 把一排学生的第一个和最后一个交换位置 |
| 06 | Normal | 复制一份班级成绩单作为备份 |
| 07 | Challenge | 传球游戏中记录每一轮球在谁手里 |
| 08 | Challenge | 队伍调头后输出新的排队顺序 |
| 09 | Challenge | 生成每天累计积分表用于查看升级进度 |

核心例子：

```lua
local scores = {80, 95, 70, 60, 100}
local n = 5

for i = 1, n do
    local current = scores[i]
end
```

### Week 08：数组统计与最值

| 项目 | 内容 |
| --- | --- |
| 核心主题 | 对一组数据做统计分析 |
| 数学基础 | 总分、平均分、最高分、最低分、分段统计 |
| 编程内容 | 数组遍历 + 条件 + 更新变量 |
| 算法内容 | 数组求和、数组计数、最大值、最小值、最大值位置、第二大值 |
| 场景 | 班级成绩、跳远比赛、成绩分段、亚军成绩 |
| 目标 | 学生掌握“先假设第一个是答案，再逐个比较”的模式 |

教学任务设计：

| 序号 | 难度 | Title |
| ---: | --- | --- |
| 01 | Simple | 计算班级成绩单的总分 |
| 02 | Simple | 计算每日金币数组的平均奖励 |
| 03 | Simple | 统计成绩单中有多少人及格 |
| 04 | Normal | 找出跳远比赛中成绩最好的选手 |
| 05 | Normal | 找出成绩单中的最低分和位置 |
| 06 | Normal | 统计游戏分数中超过目标线的人数 |
| 07 | Challenge | 成绩排行榜中找出亚军分数 |
| 08 | Challenge | 找出跳远比赛中第一名和第二名的成绩差 |
| 09 | Challenge | 按成绩区间统计优秀、及格和待提高人数 |

核心例子：

```lua
local numbers = {80, 95, 70, 60, 100}
local n = 5

local maxValue = numbers[1]
local maxIndex = 1

for i = 2, n do
    if numbers[i] > maxValue then
        maxValue = numbers[i]
        maxIndex = i
    end
end
```

### Week 09：查找与二分查找

| 项目 | 内容 |
| --- | --- |
| 核心主题 | 目标在不在，在哪里，如何更快找到 |
| 数学基础 | 数轴、区间、大小比较、范围缩小 |
| 编程内容 | 线性查找、`found`、`position`、`left/right/mid` |
| 算法内容 | 是否存在、查找位置、找第一个、找最后一个、二分查找 |
| 场景 | 号码牌、满分查找、补考名单、图书编号、查找效率比较 |
| 目标 | 学生理解线性查找和二分查找的区别 |

教学任务设计：

| 序号 | 难度 | Title |
| ---: | --- | --- |
| 01 | Simple | 在号码牌中查找指定学生是否存在 |
| 02 | Simple | 在成绩单中查找是否有人拿到满分 |
| 03 | Simple | 在背包格子中查找目标物品数量 |
| 04 | Normal | 在补考名单中找到第一个不及格学生的位置 |
| 05 | Normal | 在比赛成绩中找到最后一个满分位置 |
| 06 | Normal | 在任务列表中查找第一个未完成任务编号 |
| 07 | Challenge | 在有序图书编号中快速找到目标图书 |
| 08 | Challenge | 比较线性查找和二分查找需要检查多少次 |
| 09 | Challenge | 在有序成绩表中查找目标分数的位置 |

线性查找：

```lua
local numbers = {3, 8, 12, 7, 9}
local n = 5
local target = 12

local position = 0

for i = 1, n do
    if numbers[i] == target then
        position = i
    end
end
```

二分查找：

```lua
local numbers = {1, 3, 5, 7, 9, 12, 18}
local n = 7
local target = 9

local left = 1
local right = n
local position = 0

while left <= right and position == 0 do
    local mid = (left + right) // 2

    if numbers[mid] == target then
        position = mid
    elseif target < numbers[mid] then
        right = mid - 1
    else
        left = mid + 1
    end
end
```

### Week 10：排序算法一

| 项目 | 内容 |
| --- | --- |
| 核心主题 | 用比较和交换整理一组数据 |
| 数学基础 | 比较大小、从小到大、从大到小、排队 |
| 编程内容 | 交换数组元素、双重循环 |
| 算法内容 | 选择排序、冒泡排序第一轮、完整冒泡排序选讲 |
| 场景 | 换座位、身高排队、成绩排行榜、排序工作量 |
| 目标 | 学生理解排序的三个动作：比较、记录位置、交换 |

教学任务设计：

| 序号 | 难度 | Title |
| ---: | --- | --- |
| 01 | Simple | 按身高给一排学生从矮到高排队 |
| 02 | Simple | 把游戏分数从低到高整理 |
| 03 | Simple | 把跑步成绩从快到慢整理 |
| 04 | Normal | 用选择排序整理游戏分数排行榜 |
| 05 | Normal | 用选择排序给图书编号从小到大排序 |
| 06 | Normal | 用选择排序找出每轮剩余数据中的最小值 |
| 07 | Challenge | 统计冒泡排序整理座位表时发生了多少次交换 |
| 08 | Challenge | 用冒泡排序把成绩从高到低排列 |
| 09 | Challenge | 比较选择排序和冒泡排序的交换次数 |

选择排序：

```lua
local numbers = {5, 3, 8, 1, 2}
local n = 5

for i = 1, n - 1 do
    local minIndex = i

    for j = i + 1, n do
        if numbers[j] < numbers[minIndex] then
            minIndex = j
        end
    end

    local temp = numbers[i]
    numbers[i] = numbers[minIndex]
    numbers[minIndex] = temp
end
```

### Week 11：排序算法二与排序应用

| 项目 | 内容 |
| --- | --- |
| 核心主题 | 同一个排序问题可以有不同方法 |
| 数学基础 | 排名、中位数、重复数据、合并数据 |
| 编程内容 | 插入排序、排序后处理 |
| 算法内容 | 插入排序、中位数、去重统计、合并有序数组、并列名次 |
| 场景 | 插队排队、扑克牌整理、前三名、两个班成绩合并、并列名次 |
| 目标 | 学生能区分选择排序、冒泡排序、插入排序的思想 |

教学任务设计：

| 序号 | 难度 | Title |
| ---: | --- | --- |
| 01 | Simple | 像整理扑克牌一样插入一个新分数 |
| 02 | Simple | 在已排好队伍中插入一名新同学 |
| 03 | Simple | 找出排序后成绩单中的中间分数 |
| 04 | Normal | 合并两个班已经排好序的成绩单 |
| 05 | Normal | 排序后找出成绩前三名 |
| 06 | Normal | 统计排序后有多少个不同的分数 |
| 07 | Challenge | 处理成绩排行榜中的并列名次 |
| 08 | Challenge | 合并两个有序图书编号表并保持有序 |
| 09 | Challenge | 排序后去掉重复编号并统计剩余数量 |

插入排序：

```lua
local numbers = {5, 3, 8, 1, 2}
local n = 5

for i = 2, n do
    local key = numbers[i]
    local j = i - 1

    while j >= 1 and numbers[j] > key do
        numbers[j + 1] = numbers[j]
        j = j - 1
    end

    numbers[j + 1] = key
end
```

### Week 12：综合数学算法项目

| 项目 | 内容 |
| --- | --- |
| 核心主题 | 把算法放回真实数学场景 |
| 数学基础 | 统计、排序、查找、数量关系、几何面积 |
| 编程内容 | 综合使用变量、判断、循环、数组、函数 |
| 算法内容 | 统计、查找、排序、排名、枚举、最优方案 |
| 场景 | 成绩系统、排行榜、购物方案、骰子可能性、长方形比较 |
| 目标 | 学生能完成一个小型综合算法项目 |

教学任务设计：

| 序号 | 难度 | Title |
| ---: | --- | --- |
| 01 | Simple | 班级成绩分析系统：总分、平均分和及格人数 |
| 02 | Simple | 游戏积分分析系统：最高分、最低分和平均分 |
| 03 | Simple | 每日存钱记录系统：总额、最高单日和达标判断 |
| 04 | Normal | 游戏排行榜系统：排序、前三名和入榜判断 |
| 05 | Normal | 图书编号管理系统：查找、排序和分类统计 |
| 06 | Normal | 运动会成绩系统：最高成绩、排名和分段统计 |
| 07 | Challenge | 购物方案系统：在预算内找出最优买法 |
| 08 | Challenge | 鸡兔同笼方案系统：枚举并验证所有可能答案 |
| 09 | Challenge | 综合成绩排名系统：排序、并列名次和查找学生 |

## 7. Week 12 项目方向

### 项目 A：成绩分析系统

固定数据：

```lua
local scores = {80, 95, 67, 100, 58, 72}
local n = 6
```

功能：

```text
求总分
求平均分
找最高分
找最低分
统计及格人数
统计优秀人数
查找是否有满分
从高到低排序
计算前三名
处理并列名次，选讲
```

### 项目 B：游戏排行榜系统

固定数据：

```lua
local points = {120, 300, 250, 180, 400}
local n = 5
```

功能：

```text
找最高分
找最低分
统计超过某个分数线的人数
排序
找前三名
判断某个分数能否进入前三
查找某个分数的排名
```

### 项目 C：购物方案系统

固定数据：

```lua
local priceA = 3
local priceB = 5
local budget = 30
```

功能：

```text
统计刚好花完预算的买法
统计不超过预算的买法
找购买数量最多的方案
找剩余金额最少的方案
```

### 项目 D：骰子可能性统计

固定数据：

```lua
local diceMax = 6
```

功能：

```text
统计两个骰子的所有点数和
统计每个点数和出现次数
找出现次数最多的点数和
比较不同点数和的可能性
```

### 项目 E：长方形面积比较

固定数据：

```lua
local lengths = {5, 8, 6, 10}
local widths = {4, 3, 7, 2}
local n = 4
```

功能：

```text
计算每个长方形面积
找面积最大的长方形
找面积最小的长方形
按面积排序，选讲
```

## 8. 核心算法覆盖清单

| 编号 | 算法 |
| ---: | --- |
| 01 | 顺序计算 |
| 02 | 变量更新 |
| 03 | 变量交换 |
| 04 | 条件分类 |
| 05 | 奇偶判断 |
| 06 | 倍数判断 |
| 07 | 计数算法 |
| 08 | 分类统计 |
| 09 | 累加算法 |
| 10 | 平均值 |
| 11 | 累计数组 / 前缀和雏形 |
| 12 | 因数枚举 |
| 13 | 质数判断 |
| 14 | 最大公因数枚举 |
| 15 | 最小公倍数枚举 |
| 16 | 单变量枚举 |
| 17 | 双变量枚举 |
| 18 | 数组遍历 |
| 19 | 数组复制 |
| 20 | 数组反转 |
| 21 | 数组右移 |
| 22 | 数组求和 |
| 23 | 数组计数 |
| 24 | 最大值 / 最小值 |
| 25 | 最大值位置 |
| 26 | 第二大值 |
| 27 | 线性查找 |
| 28 | 二分查找 |
| 29 | 选择排序 |
| 30 | 冒泡排序 |
| 31 | 插入排序 |
| 32 | 排序后找前三名 |
| 33 | 排序后去重 |
| 34 | 合并有序数组 |
| 35 | 并列排名 |
| 36 | 综合项目建模 |

## 9. 对应小学数学内容

| 数学领域 | 课程对应内容 |
| --- | --- |
| 数与运算 | 四则运算、倍数、因数、质数、余数、平均数 |
| 数量关系 | 分糖、购物、鸡兔同笼、组合枚举 |
| 统计与概率 | 成绩统计、分段统计、骰子可能性、频数 |
| 图形与几何 | 长方形面积、周长、面积比较 |
| 综合与实践 | 成绩系统、排行榜、购物方案、骰子分析 |

## 10. Lesson 与题目命名规则

Lesson 命名规则：

```text
Lesson 序号：固定编号
Lesson 名称：Week YY - 描述
```

示例：

```text
Lesson 01
Lesson 名称：Week 01 - Variables

Lesson 02
Lesson 名称：Week 01 - 变量
```

题目字段规则：

```text
序号
难度
Title
```

题目示例：

| 序号 | 难度 | Title |
| ---: | --- | --- |
| 01 | Simple | 游戏角色受到攻击后还剩多少血量 |
| 02 | Simple | 文具店买铅笔一共要花多少钱 |
| 03 | Simple | 操场长方形区域的面积是多少 |
| 04 | Normal | 打怪获得金币后背包里共有多少金币 |
| 05 | Normal | 买铅笔和橡皮后还剩多少钱 |
| 06 | Normal | 跑步训练后累计跑了多少米 |
| 07 | Challenge | 两个同学换座位后座位编号如何交换 |
| 08 | Challenge | 游戏角色先受伤再回血后剩多少血量 |
| 09 | Challenge | 商店打包购买后计算总价和找零 |

每套 lesson 结构固定：

```text
Simple 01-03
Normal 04-06
Challenge 07-09
```

## 11. 后续生成 Lesson 的标准模板

每套 lesson 建议包含以下字段：

```text
Lesson Number
Lesson Name
Language
Week
Objectives
Limits
Concepts
Questions
Review
```

每道题只使用以下索引字段：

```text
序号
难度
Title
```

题目详情可在索引字段下继续包含：

```text
Story
Task
Fixed Data
Expected Output / Result
Hints
Starter Code
Reference Solution
```

每套 lesson 建议包含以下部分：

```text
1. 本课目标
2. 本课限制
3. 核心概念
4. 9 道场景化题目
5. 本课复盘
```

这份教学文档可作为后续生成 24 套 lesson 的统一骨架。
