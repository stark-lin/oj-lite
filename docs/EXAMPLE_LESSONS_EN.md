# Lua Basic Algorithms Bilingual Course Teaching Document

## Document Purpose

This document defines a Lua basic algorithms course for students who have completed Scratch and are ready to move into text-based programming. The course is designed as a bilingual lesson set. For the same week, the English lesson uses an odd lesson number, and the Chinese lesson uses an even lesson number.

Each lesson number is fixed and should be used for ordering and reference. The lesson name should only contain the week number and the topic description. Language, difficulty, and question metadata should not be included in the lesson name.

Numbering rule:

```text
Lesson 01, 03, 05, ... 23: English lesson
Lesson 02, 04, 06, ... 24: Chinese lesson
```

## 1. Course Positioning

This course is for students who have finished Scratch and are preparing to learn text-based programming. Lua is used as a lightweight language for expressing algorithms. The main goal is to build foundational algorithmic thinking for later Python / C / C++ learning.

Course principles:

| Principle | Description |
| --- | --- |
| Algorithm first | Lua is used as an expression tool. The course is not designed as a full Lua syntax course. |
| Simple progression | Each lesson introduces only a small number of new ideas. Questions stay short, clear, and testable. |
| Scenario-based | All questions are built from familiar student contexts such as games, school, shopping, sports, and scores. |
| Bilingual pairing | Each week has one English lesson and one Chinese lesson with aligned structure and concepts. |
| Fixed structure | Each lesson is divided into Simple, Normal, and Challenge levels. |

## 2. Technical Boundaries

Do not use:

```lua
io.read
string.*
table.insert
table.sort
pairs
ipairs
math.*
```

Only use:

```text
local variables
if / elseif / else
for
while
function
array tables
array indexes
basic arithmetic
comparison operators
logical operators
```

## 3. Overall Structure

The course contains 12 weeks. Each week has 2 lessons. For the same topic, the English lesson is placed first, followed by the Chinese lesson:

```text
English lesson
Chinese lesson
```

Total:

```text
12 weeks x 2 language lessons = 24 lessons
```

Each lesson contains 9 questions:

| Difficulty | Count | Purpose |
| --- | ---: | --- |
| Simple | 3 questions | Warm-up and concept confirmation |
| Normal | 3 questions | Main practice for the core algorithm |
| Challenge | 3 questions | Extension, integration, or homework |

Total question count:

```text
9 Chinese questions + 9 English questions per week = 18 questions per week
12 weeks = 108 Chinese questions + 108 English questions = 216 questions
```

## 4. Lesson Numbering

| Lesson No. | Language | Lesson Name |
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

## 5. 12-Week Course Track

| Week | Topic | Core Algorithm Skills |
| ---: | --- | --- |
| Week 01 | Variables | Sequential calculation, variable updates, variable swapping |
| Week 02 | Conditions and Loops | Conditional classification, loop counting, boundary checks |
| Week 03 | Counting and Classification | Conditional counting, classification statistics |
| Week 04 | Sum, Average, and Accumulation | Summation, conditional summation, averages, accumulation |
| Week 05 | Factors, Multiples, and Prime Numbers | Enumeration, factor counting, prime checking |
| Week 06 | Quantity Relations and Double Enumeration | Nested loops, solution search, combination counting |
| Week 07 | Arrays and Table Data | Array access, modification, traversal, copying |
| Week 08 | Array Statistics and Extremes | Array sum, array count, maximum, minimum, second largest value |
| Week 09 | Search and Binary Search | Linear search, position search, binary search |
| Week 10 | Sorting Algorithms I | Selection sort, bubble sort |
| Week 11 | Sorting Algorithms II and Applications | Insertion sort, post-sort processing, merging sorted arrays |
| Week 12 | Integrated Math Algorithm Projects | Statistics, search, sorting, enumeration, modeling |

## 6. Weekly Teaching Outline

### Week 01: Variables

| Item | Content |
| --- | --- |
| Core theme | A variable is a box that stores data. |
| Math foundation | Four operations, area, perimeter, total price |
| Programming content | `local`, assignment, variable updates, temporary variables |
| Algorithm content | Sequential calculation, variable swapping, state updates |
| Scenarios | Game health, playground area, stationery shopping, coin changes, seat swapping |
| Goal | Students can use variables to express the calculation process of a concrete word problem. |

Teaching task design:

| No. | Difficulty | Title |
| ---: | --- | --- |
| 01 | Simple | How Much Health Is Left After a Game Character Is Attacked |
| 02 | Simple | How Much Do the Pencils Cost at the Stationery Store |
| 03 | Simple | What Is the Area of the Rectangular Playground |
| 04 | Normal | How Many Coins Are in the Bag After Defeating a Monster |
| 05 | Normal | How Much Money Is Left After Buying Pencils and Erasers |
| 06 | Normal | How Many Meters Have Been Run After Training |
| 07 | Challenge | How Do Two Students Swap Seat Numbers |
| 08 | Challenge | How Much Health Is Left After Damage and Healing |
| 09 | Challenge | Calculate Total Price and Change for a Store Bundle |

Core example:

```lua
local hp = 100
local damage = 25
local remainHp = hp - damage
```

### Week 02: Conditions and Loops

| Item | Content |
| --- | --- |
| Core theme | Conditions choose branches, and loops handle repetition. |
| Math foundation | Comparisons, odd and even numbers, multiples, boundary values |
| Programming content | `if`, `else`, `elseif`, `for`, `while` |
| Algorithm content | Conditional classification, loop counting, boundary checks |
| Scenarios | Temperature checks, student numbers, passing scores, monster health, leap years |
| Goal | Students can write basic "if..." and "repeat..." algorithm structures. |

Teaching task design:

| No. | Difficulty | Title |
| ---: | --- | --- |
| 01 | Simple | Check Whether an Exam Score Reaches the Passing Line |
| 02 | Simple | Check Whether a Student Number Is Odd or Even |
| 03 | Simple | Check Whether the Temperature Is Good for Outdoor Sports |
| 04 | Normal | How Many Rounds Until a Monster Is Defeated |
| 05 | Normal | Count How Many Even Numbers There Are from 1 to n |
| 06 | Normal | Check Whether a Number Is a Multiple of 3 |
| 07 | Challenge | Check Whether the School Sports Day Falls in a Leap Year |
| 08 | Challenge | Choose a Drink in a Vending Machine Based on Money |
| 09 | Challenge | Decide How Many Stars a Game Level Gets Based on Score |

Core example:

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

### Week 03: Counting and Classification

| Item | Content |
| --- | --- |
| Core theme | How many items meet a condition? |
| Math foundation | Counting, classification, multiples, remainders |
| Programming content | Loop + condition + `count` |
| Algorithm content | Counting algorithm, conditional counting, classification statistics |
| Scenarios | Seat numbers, student numbers, task numbers, library levels, lottery numbers |
| Goal | Students understand when to use `count = count + 1`. |

Teaching task design:

| No. | Difficulty | Title |
| ---: | --- | --- |
| 01 | Simple | Count How Many Even Seat Numbers Are in a Class |
| 02 | Simple | Count Multiples of 5 from 1 to 50 |
| 03 | Simple | Count How Many Odd Task Numbers There Are |
| 04 | Normal | Count How Many Lucky Lottery Numbers There Are |
| 05 | Normal | Count Regular Books by Library Number |
| 06 | Normal | Count Athletes Whose Numbers Are Divisible by 3 |
| 07 | Challenge | Classify Library Books into Three Levels by Number |
| 08 | Challenge | Count Game Tasks by Difficulty Number |
| 09 | Challenge | Group Student Numbers by Remainder and Count Each Group |

Core example:

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

### Week 04: Sum, Average, and Accumulation

| Item | Content |
| --- | --- |
| Core theme | What is the total, what is the average, and where does the accumulation reach? |
| Math foundation | Repeated addition, averages, accumulated quantities |
| Programming content | `sum`, loop-based summation, early idea of accumulated arrays |
| Algorithm content | Summation, conditional summation, average, accumulation table, prefix-sum foundation |
| Scenarios | Stair numbers, level rewards, daily coins, saving goals |
| Goal | Students can distinguish the use of `count` and `sum`. |

Teaching task design:

| No. | Difficulty | Title |
| ---: | --- | --- |
| 01 | Simple | Calculate the Total Coin Reward for One Week |
| 02 | Simple | Calculate the Sum of Stair Numbers from 1 to n |
| 03 | Simple | Calculate the Total Distance Run After Several Days |
| 04 | Normal | Calculate the Average Score of Several Class Quizzes |
| 05 | Normal | Calculate Total Savings on Day n |
| 06 | Normal | Calculate Total and Average Coin Rewards from Game Levels |
| 07 | Challenge | Find the First Day When Savings Reach the Toy Goal |
| 08 | Challenge | Find the First Floor Where Total Steps Exceed a Target |
| 09 | Challenge | Find the Day When Daily Game Rewards Reach an Upgrade Goal |

Core example:

```lua
local n = 100
local sum = 0

for i = 1, n do
    sum = sum + i
end
```

### Week 05: Factors, Multiples, and Prime Numbers

| Item | Content |
| --- | --- |
| Core theme | Use enumeration to solve number property problems. |
| Math foundation | Factors, multiples, prime numbers, common factors, common multiples |
| Programming content | `%`, loop enumeration, Boolean variables |
| Algorithm content | Factor enumeration, factor counting, prime checking, greatest common factor enumeration, least common multiple enumeration |
| Scenarios | Team grouping, puzzle blocks, special numbers, running meetups |
| Goal | Students understand the algorithmic idea of trying possibilities one by one. |

Teaching task design:

| No. | Difficulty | Title |
| ---: | --- | --- |
| 01 | Simple | Check Whether a Team Can Be Split into Equal Groups |
| 02 | Simple | Check Whether One Number Is a Multiple of Another |
| 03 | Simple | Check Whether Puzzle Blocks Can Be Shared Equally |
| 04 | Normal | Find All Ways to Split a Set of Puzzle Blocks |
| 05 | Normal | Count How Many Factors a Special Number Has |
| 06 | Normal | Find When Two Running Cycles Meet for the First Time |
| 07 | Challenge | Check Whether a Special Number Is Prime |
| 08 | Challenge | Find the Largest Common Group Size for Two Classes |
| 09 | Challenge | Find the First Minute When Two Alarms Ring Together |

Core example:

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

### Week 06: Quantity Relations and Double Enumeration

| Item | Content |
| --- | --- |
| Core theme | How can we find an answer when two variables change together? |
| Math foundation | Sum and difference relations, multiplication relations, chicken and rabbit problems, shopping plans |
| Programming content | Nested loops, conditional filtering |
| Algorithm content | Double-variable enumeration, combination counting, solution search |
| Scenarios | Candy buying, rectangle puzzles, dice games, stationery shopping, chicken and rabbit problems |
| Goal | Students can use nested loops to solve simple word problems. |

Teaching task design:

| No. | Difficulty | Title |
| ---: | --- | --- |
| 01 | Simple | How Many Ways Are There to Buy 20 Candies of Two Types |
| 02 | Simple | How Many Dice Outcomes Have a Target Sum |
| 03 | Simple | How Many Length and Width Pairs Match a Fixed Rectangle Perimeter |
| 04 | Normal | Count Stationery Plans for Buying Pencils and Erasers with a Fixed Budget |
| 05 | Normal | Count Game Shop Plans That Spend All Coins on Potions and Shields |
| 06 | Normal | Count Ways to Split a Fixed Signup Quota Between Two Classes |
| 07 | Challenge | Chicken and Rabbit Problem: Find Animal Counts from Heads and Legs |
| 08 | Challenge | Buy Child and Adult Tickets with an Exact Budget |
| 09 | Challenge | Find the Shopping Plan with the Most Items Under a Budget |

Core example:

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

### Week 07: Arrays and Table Data

| Item | Content |
| --- | --- |
| Core theme | How can a group of data be stored and processed? |
| Math foundation | Table data, sequence numbers, positions |
| Programming content | Array tables, indexes, traversal, modification |
| Algorithm content | Array access, array modification, array traversal, array copying, accumulated arrays |
| Scenarios | Score sheets, fitness records, daily points, reversing a line, passing games |
| Goal | Students understand an array as a row of numbered boxes. |

Teaching task design:

| No. | Difficulty | Title |
| ---: | --- | --- |
| 01 | Simple | Read the Score of a Student from a Score Sheet |
| 02 | Simple | Read a Student's Long Jump Result from a Fitness Table |
| 03 | Simple | Check the Item Count in a Game Bag Slot |
| 04 | Normal | Modify the Points for One Day in a Daily Points Table |
| 05 | Normal | Swap the First and Last Students in a Line |
| 06 | Normal | Copy a Class Score Sheet as a Backup |
| 07 | Challenge | Record Who Holds the Ball in Each Round of a Passing Game |
| 08 | Challenge | Output the New Order After a Team Turns Around |
| 09 | Challenge | Generate an Accumulated Points Table for Upgrade Progress |

Core example:

```lua
local scores = {80, 95, 70, 60, 100}
local n = 5

for i = 1, n do
    local current = scores[i]
end
```

### Week 08: Array Statistics and Extremes

| Item | Content |
| --- | --- |
| Core theme | Analyze a group of data. |
| Math foundation | Total score, average score, highest score, lowest score, range statistics |
| Programming content | Array traversal + conditions + variable updates |
| Algorithm content | Array sum, array count, maximum, minimum, maximum position, second largest value |
| Scenarios | Class scores, long jump competition, score ranges, runner-up score |
| Goal | Students master the pattern of assuming the first item is the answer, then comparing one by one. |

Teaching task design:

| No. | Difficulty | Title |
| ---: | --- | --- |
| 01 | Simple | Calculate the Total Score of a Class Score Sheet |
| 02 | Simple | Calculate the Average Daily Coin Reward |
| 03 | Simple | Count How Many Students Passed |
| 04 | Normal | Find the Best Player in a Long Jump Competition |
| 05 | Normal | Find the Lowest Score and Its Position |
| 06 | Normal | Count Game Scores Above a Target Line |
| 07 | Challenge | Find the Runner-Up Score in a Score Ranking |
| 08 | Challenge | Find the Difference Between First and Second Place in Long Jump |
| 09 | Challenge | Count Excellent, Passing, and Needs-Improvement Score Ranges |

Core example:

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

### Week 09: Search and Binary Search

| Item | Content |
| --- | --- |
| Core theme | Is the target present, where is it, and how can we find it faster? |
| Math foundation | Number lines, intervals, comparison, narrowing a range |
| Programming content | Linear search, `found`, `position`, `left/right/mid` |
| Algorithm content | Existence search, position search, first match, last match, binary search |
| Scenarios | Number cards, full-score search, retake lists, book numbers, search efficiency comparison |
| Goal | Students understand the difference between linear search and binary search. |

Teaching task design:

| No. | Difficulty | Title |
| ---: | --- | --- |
| 01 | Simple | Search for a Student in Number Cards |
| 02 | Simple | Check Whether Anyone Got a Full Score |
| 03 | Simple | Search for a Target Item Count in a Bag Slot |
| 04 | Normal | Find the First Failing Student in a Retake List |
| 05 | Normal | Find the Last Full-Score Position in Competition Results |
| 06 | Normal | Find the First Unfinished Task Number |
| 07 | Challenge | Quickly Find a Target Book in Sorted Book Numbers |
| 08 | Challenge | Compare How Many Checks Linear Search and Binary Search Need |
| 09 | Challenge | Find a Target Score in a Sorted Score Table |

Linear search:

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

Binary search:

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

### Week 10: Sorting Algorithms I

| Item | Content |
| --- | --- |
| Core theme | Organize a group of data using comparison and swapping. |
| Math foundation | Comparing size, ascending order, descending order, lining up |
| Programming content | Swapping array elements, nested loops |
| Algorithm content | Selection sort, first round of bubble sort, full bubble sort as an extension |
| Scenarios | Seat swapping, height ordering, score rankings, sorting workload |
| Goal | Students understand the three actions of sorting: compare, record position, swap. |

Teaching task design:

| No. | Difficulty | Title |
| ---: | --- | --- |
| 01 | Simple | Line Up Students by Height from Short to Tall |
| 02 | Simple | Organize Game Scores from Low to High |
| 03 | Simple | Organize Running Results from Fast to Slow |
| 04 | Normal | Use Selection Sort to Organize a Game Score Ranking |
| 05 | Normal | Use Selection Sort to Sort Book Numbers from Small to Large |
| 06 | Normal | Use Selection Sort to Find the Minimum in Each Remaining Round |
| 07 | Challenge | Count Swaps When Bubble Sorting a Seat Table |
| 08 | Challenge | Use Bubble Sort to Sort Scores from High to Low |
| 09 | Challenge | Compare Swap Counts of Selection Sort and Bubble Sort |

Selection sort:

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

### Week 11: Sorting Algorithms II and Applications

| Item | Content |
| --- | --- |
| Core theme | The same sorting problem can be solved in different ways. |
| Math foundation | Ranking, median, duplicate data, merging data |
| Programming content | Insertion sort, post-sort processing |
| Algorithm content | Insertion sort, median, unique counting, merging sorted arrays, tied ranking |
| Scenarios | Joining a line, organizing playing cards, top three scores, merging two class score lists, tied ranks |
| Goal | Students can distinguish the ideas behind selection sort, bubble sort, and insertion sort. |

Teaching task design:

| No. | Difficulty | Title |
| ---: | --- | --- |
| 01 | Simple | Insert a New Score Like Organizing Playing Cards |
| 02 | Simple | Insert a New Student into an Ordered Line |
| 03 | Simple | Find the Middle Score After Sorting |
| 04 | Normal | Merge Two Already Sorted Class Score Sheets |
| 05 | Normal | Find the Top Three Scores After Sorting |
| 06 | Normal | Count How Many Different Scores Remain After Sorting |
| 07 | Challenge | Handle Tied Ranks in a Score Ranking |
| 08 | Challenge | Merge Two Sorted Book Number Lists and Keep Them Sorted |
| 09 | Challenge | Remove Duplicate Numbers After Sorting and Count the Remaining Items |

Insertion sort:

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

### Week 12: Integrated Math Algorithm Projects

| Item | Content |
| --- | --- |
| Core theme | Put algorithms back into real math scenarios. |
| Math foundation | Statistics, sorting, search, quantity relations, geometric area |
| Programming content | Combine variables, conditions, loops, arrays, and functions |
| Algorithm content | Statistics, search, sorting, ranking, enumeration, optimal plans |
| Scenarios | Score systems, leaderboards, shopping plans, dice possibilities, rectangle comparison |
| Goal | Students can complete a small integrated algorithm project. |

Teaching task design:

| No. | Difficulty | Title |
| ---: | --- | --- |
| 01 | Simple | Class Score Analysis System: Total, Average, and Passing Count |
| 02 | Simple | Game Points Analysis System: Highest, Lowest, and Average |
| 03 | Simple | Daily Savings Record System: Total, Highest Day, and Goal Check |
| 04 | Normal | Game Leaderboard System: Sorting, Top Three, and Entry Check |
| 05 | Normal | Book Number Management System: Search, Sorting, and Classification |
| 06 | Normal | Sports Day Results System: Best Result, Ranking, and Range Statistics |
| 07 | Challenge | Shopping Plan System: Find the Best Plan Within a Budget |
| 08 | Challenge | Chicken and Rabbit System: Enumerate and Verify All Possible Answers |
| 09 | Challenge | Integrated Score Ranking System: Sorting, Tied Ranking, and Student Search |

## 7. Week 12 Project Directions

### Project A: Score Analysis System

Fixed data:

```lua
local scores = {80, 95, 67, 100, 58, 72}
local n = 6
```

Functions:

```text
Calculate total score
Calculate average score
Find the highest score
Find the lowest score
Count passing students
Count excellent students
Check whether there is a full score
Sort from high to low
Calculate the top three
Handle tied ranks as an extension
```

### Project B: Game Leaderboard System

Fixed data:

```lua
local points = {120, 300, 250, 180, 400}
local n = 5
```

Functions:

```text
Find the highest score
Find the lowest score
Count how many players exceed a score line
Sort scores
Find the top three
Check whether a score can enter the top three
Find the rank of a score
```

### Project C: Shopping Plan System

Fixed data:

```lua
local priceA = 3
local priceB = 5
local budget = 30
```

Functions:

```text
Count plans that spend the exact budget
Count plans that do not exceed the budget
Find the plan with the largest item count
Find the plan with the smallest remaining money
```

### Project D: Dice Possibility Statistics

Fixed data:

```lua
local diceMax = 6
```

Functions:

```text
Count all possible sums of two dice
Count how many times each sum appears
Find the sum that appears most often
Compare the possibilities of different sums
```

### Project E: Rectangle Area Comparison

Fixed data:

```lua
local lengths = {5, 8, 6, 10}
local widths = {4, 3, 7, 2}
local n = 4
```

Functions:

```text
Calculate each rectangle's area
Find the rectangle with the largest area
Find the rectangle with the smallest area
Sort by area as an extension
```

## 8. Core Algorithm Coverage

| No. | Algorithm |
| ---: | --- |
| 01 | Sequential calculation |
| 02 | Variable update |
| 03 | Variable swapping |
| 04 | Conditional classification |
| 05 | Odd and even checking |
| 06 | Multiple checking |
| 07 | Counting algorithm |
| 08 | Classification statistics |
| 09 | Summation algorithm |
| 10 | Average value |
| 11 | Accumulated array / prefix-sum foundation |
| 12 | Factor enumeration |
| 13 | Prime checking |
| 14 | Greatest common factor enumeration |
| 15 | Least common multiple enumeration |
| 16 | Single-variable enumeration |
| 17 | Double-variable enumeration |
| 18 | Array traversal |
| 19 | Array copying |
| 20 | Array reversal |
| 21 | Array right shift |
| 22 | Array summation |
| 23 | Array counting |
| 24 | Maximum / minimum |
| 25 | Maximum position |
| 26 | Second largest value |
| 27 | Linear search |
| 28 | Binary search |
| 29 | Selection sort |
| 30 | Bubble sort |
| 31 | Insertion sort |
| 32 | Finding top three after sorting |
| 33 | Removing duplicates after sorting |
| 34 | Merging sorted arrays |
| 35 | Tied ranking |
| 36 | Integrated project modeling |

## 9. Primary Math Alignment

| Math Area | Course Alignment |
| --- | --- |
| Numbers and operations | Four operations, multiples, factors, prime numbers, remainders, averages |
| Quantity relations | Candy sharing, shopping, chicken and rabbit problems, combination enumeration |
| Statistics and probability | Score statistics, range statistics, dice possibilities, frequencies |
| Geometry | Rectangle area, perimeter, area comparison |
| Integrated practice | Score systems, leaderboards, shopping plans, dice analysis |

## 10. Lesson and Question Naming Rules

Lesson naming rule:

```text
Lesson number: fixed number
Lesson name: Week YY - Description
```

Examples:

```text
Lesson 01
Lesson name: Week 01 - Variables

Lesson 02
Lesson name: Week 01 - 变量
```

Question field rule:

```text
No.
Difficulty
Title
```

Question example:

| No. | Difficulty | Title |
| ---: | --- | --- |
| 01 | Simple | How Much Health Is Left After a Game Character Is Attacked |
| 02 | Simple | How Much Do the Pencils Cost at the Stationery Store |
| 03 | Simple | What Is the Area of the Rectangular Playground |
| 04 | Normal | How Many Coins Are in the Bag After Defeating a Monster |
| 05 | Normal | How Much Money Is Left After Buying Pencils and Erasers |
| 06 | Normal | How Many Meters Have Been Run After Training |
| 07 | Challenge | How Do Two Students Swap Seat Numbers |
| 08 | Challenge | How Much Health Is Left After Damage and Healing |
| 09 | Challenge | Calculate Total Price and Change for a Store Bundle |

Each lesson uses this fixed structure:

```text
Simple 01-03
Normal 04-06
Challenge 07-09
```

## 11. Standard Template for Future Lesson Generation

Each lesson should contain these fields:

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

Each question should use only these index fields:

```text
No.
Difficulty
Title
```

Question details may be included under the index fields:

```text
Story
Task
Fixed Data
Expected Output / Result
Hints
Starter Code
Reference Solution
```

Each lesson should contain these sections:

```text
1. Lesson Objectives
2. Lesson Limits
3. Core Concepts
4. 9 Scenario-Based Questions
5. Lesson Review
```

This teaching document can be used as the shared structure for generating all 24 lessons.
