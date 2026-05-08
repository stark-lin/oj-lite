# Lua Basic Algorithms Bilingual Course Teaching Document

## 1. Document Purpose

This document plans a Lua basic algorithms course for students who have advanced beyond Scratch and are preparing to move into text-based programming. The course uses paired Chinese and English lessons. For the same week, the English lesson uses an odd lesson number, and the Chinese lesson uses an even lesson number.

Each lesson number is fixed for ordering and reference. Lesson names use only the week number and topic description; language and difficulty metadata are not written into the lesson name.

Numbering rule:

```text
Lesson 01, 03, 05, ... 23: English lesson
Lesson 02, 04, 06, ... 24: Chinese lesson
```

---

## 2. Course Positioning

This course is for students who have finished Scratch and are preparing to learn text-based programming. Lua is used as a lightweight language for expressing algorithms. The focus is foundational algorithmic thinking that prepares students for later Python / C / C++ learning.

| Principle | Description |
| --- | --- |
| Algorithm first | Lua is only the expression tool; the course is not designed as a full Lua syntax course. |
| Simple progression | Each lesson introduces only a few new ideas, and problems remain short, clear, and testable. |
| Scenario-based | Problems come from familiar student contexts such as games, school, shopping, sports, and scores. |
| Bilingual pairing | Each week has one Chinese lesson and one English lesson with aligned structure and concepts. |
| Fixed structure | Each lesson is divided into Simple, Normal, and Challenge levels. |

---

## 3. Technical Boundaries

### Do Not Use

```lua
io.read
string.*
table.insert
table.sort
pairs
ipairs
math.*
```

### Only Use

```lua
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

---

## 4. Overall Structure

The course has 12 weeks. Each week has 2 language lessons.

```text
12 weeks x 2 language lessons = 24 lessons
```

Each lesson has exactly 9 questions:

| Difficulty | Count | Purpose |
| --- | --: | --- |
| Simple | 3 questions | Warm-up and confirmation of the week's core concept |
| Normal | 3 questions | Main practice, with appropriate review of earlier knowledge |
| Challenge | 3 questions | Integrated application with clearly higher difficulty |

Total question count:

```text
9 Chinese questions + 9 English questions per week = 18 questions
12 weeks = 108 Chinese questions + 108 English questions = 216 questions
```

---

## 5. Lesson Numbering Order

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

---

## 6. Question Difficulty Design Rules

Starting from Week 02, difficulty progresses as:

```text
This week's practice -> previous knowledge combination -> integrated challenge
```

| Difficulty | Design Standard |
| --- | --- |
| Simple | Checks only the week's core concept and avoids introducing multiple old knowledge points at the same time. |
| Normal | May combine variables, conditions, loops, counting, accumulation, or array operations from earlier weeks, but the goal still centers on the week's topic. |
| Challenge | Must be integrated and usually requires 2 to 4 algorithm ideas at the same time. |
| Hint | Gives only a direction, not key conditions, loop ranges, update formulas, or complete code. |

---

## 7. Hint Control Rules

### Simple Hint

May suggest which structure to use.

```text
Hint: Check the condition first, then use a variable to record the result.
```

Avoid giving the complete condition directly.

### Normal Hint

May suggest solution steps, but should not give the complete condition expression.

```text
Hint: Calculate the current state first, then classify it from that state.
```

### Challenge Hint

Only suggests how to split the problem.

```text
Hint: Split the problem into enumerating plans, checking validity, and updating the best answer.
```

Avoid directly giving loop variables, conditions, or complete formulas.

---

## 8. 12-Week Course Track

| Week | Topic | Core Algorithm Skills |
| ---: | --- | --- |
| Week 01 | Variables | Sequential calculation, variable swapping, state updates, multi-step settlement |
| Week 02 | Conditions and Loops | Conditional classification, loop counting, boundary checks, state updates |
| Week 03 | Counting and Classification | Counting algorithms, conditional counting, classification statistics |
| Week 04 | Sum, Average, and Accumulation | Summation, conditional summation, averages, target accumulation |
| Week 05 | Factors, Multiples, and Prime Numbers | Factor enumeration, factor counting, prime checks, greatest common factor enumeration, least common multiple enumeration |
| Week 06 | Quantity Relations and Double Enumeration | Double-variable enumeration, combination counting, plan search, best-plan update |
| Week 07 | Arrays and Table Data | Array access, array modification, array traversal, array copying, accumulated arrays |
| Week 08 | Array Statistics and Extremes | Array summation, array counting, maximum value, minimum value, maximum-value position, second-largest value |
| Week 09 | Search and Binary Search | Existence search, position search, first-position search, last-position search, binary search |
| Week 10 | Sorting Algorithms I | Selection sort, bubble sort, swap-count statistics |
| Week 11 | Sorting Algorithms II and Applications | Insertion sort, middle values, deduplication statistics, merging sorted arrays, tied ranks |
| Week 12 | Integrated Math Algorithm Projects | Statistics, search, sorting, ranking, enumeration, best-plan search |

---

# 9. Detailed Weekly Outline

---

## Week 01: Variables

| Item | Content |
| --- | --- |
| Core theme | Variables are boxes for storing data and intermediate results |
| Math foundation | Arithmetic, area, total price, average, difference, change, modulo |
| Programming content | `local`, assignment, variable updates, temporary variables, operator precedence |
| Algorithm content | Sequential calculation, variable swapping, state updates, multi-step settlement |
| Scenarios | Game health, playground area, stationery shopping, coin changes, running accumulation, seat swapping, bundle discount settlement |
| Goal | Students can use variables to break down real word problems and express complete calculations with addition, subtraction, multiplication, division, and modulo. |

### Teaching Task Design

| No. | Difficulty | Title |
| --: | --- | --- |
| 01 | Simple | How Much Health Is Left After a Game Character Is Attacked |
| 02 | Simple | How Much Do the Pencils Cost at the Stationery Store |
| 03 | Simple | What Is the Area of the Rectangular Playground |
| 04 | Normal | How Many Coins Are in the Bag After Defeating a Monster |
| 05 | Normal | How Much Money Is Left After Buying Pencils and Erasers |
| 06 | Normal | How Many Meters Have Been Run After Training |
| 07 | Challenge | Update the Seat Statistics Card After Two Students Swap Seats |
| 08 | Challenge | Final Health Settlement After Consecutive Battles |
| 09 | Challenge | Store Bundle Discount, Packaging Fee, and Change Settlement |

### Progression Notes

| Difficulty | Combination Method |
| --- | --- |
| Simple | Only practices direct variable calculation. |
| Normal | Combines variable updates and multi-step arithmetic. |
| Challenge | Combines swapping, state settlement, modulo, and multi-step calculation. |

---

## Week 02: Conditions and Loops

| Item | Content |
| --- | --- |
| Core theme | Conditions choose branches, and loops handle repetition |
| Math foundation | Comparison, odd and even numbers, multiples, boundary values |
| Programming content | `if`, `else`, `elseif`, `for`, `while` |
| Algorithm content | Conditional classification, loop counting, boundary checks, state updates |
| Scenarios | Exam passing, student numbers, game battles, shopping checks, vending machines, level star ratings |
| Goal | Students can write the two basic algorithm structures: if something is true, do this; and repeat this process. |

### Teaching Task Design

| No. | Difficulty | Title |
| --: | --- | --- |
| 01 | Simple | Check Whether an Exam Score Reaches the Passing Line |
| 02 | Simple | Check Whether a Student Number Is Odd or Even |
| 03 | Simple | Use a Loop to Count All Numbers from 1 to n |
| 04 | Normal | Check Whether a Game Character Is Still Alive After an Attack |
| 05 | Normal | Calculate How Many Rounds Until a Monster Is Defeated |
| 06 | Normal | Check Whether an Eraser Can Still Be Bought After Shopping |
| 07 | Challenge | Rating Settlement After Consecutive Game Battles |
| 08 | Challenge | Final Result After Consecutive Vending Machine Purchases |
| 09 | Challenge | Star Rating After Consecutive Game Level Scores |

### Progression Notes

| Difficulty | Combination Method |
| --- | --- |
| Simple | Practices only condition checks, odd-even checks, and basic loops. |
| Normal | Combines Week 01 variable calculation and variable updates. |
| Challenge | Combines loops, state updates, multiple conditions, and final classification. |

---

## Week 03: Counting and Classification

| Item | Content |
| --- | --- |
| Core theme | How many items satisfy a condition |
| Math foundation | Counting, classification, multiples, remainders |
| Programming content | Loop + condition + `count` |
| Algorithm content | Counting algorithms, conditional counting, classification statistics |
| Scenarios | Seat numbers, student numbers, task numbers, book grouping, lottery numbers |
| Goal | Students understand when to use `count = count + 1`. |

### Teaching Task Design

| No. | Difficulty | Title |
| --: | --- | --- |
| 01 | Simple | Count How Many Even Numbers There Are from 1 to n |
| 02 | Simple | Count How Many Multiples of 5 There Are from 1 to n |
| 03 | Simple | Count How Many Odd Task Numbers There Are |
| 04 | Normal | Count Numbers That Are Even and Greater Than the Target |
| 05 | Normal | Count Game Levels That Can Receive Reward Coins |
| 06 | Normal | Count Numbers Divisible by 3 but Not Even |
| 07 | Challenge | Split Student Numbers into Three Remainder Groups and Count Them |
| 08 | Challenge | Count Normal, Hard, and Hidden Game Tasks by Number Rules |
| 09 | Challenge | Compare How Many Numbers Two Reward Rules Select |

### Progression Notes

| Difficulty | Combination Method |
| --- | --- |
| Simple | Practices single-condition counting. |
| Normal | Combines multiple conditions from Week 02. |
| Challenge | Uses multiple counters, classification priority, and result comparison. |

---

## Week 04: Sum, Average, and Accumulation

| Item | Content |
| --- | --- |
| Core theme | What is the total, what is the average, and when is the target reached |
| Math foundation | Repeated addition, averages, accumulated quantities |
| Programming content | `sum`, loop accumulation, accumulated variables |
| Algorithm content | Summation, conditional summation, averages, target accumulation |
| Scenarios | Stair numbers, level rewards, daily coins, savings goals |
| Goal | Students can distinguish the roles of `count` and `sum`. |

### Teaching Task Design

| No. | Difficulty | Title |
| --: | --- | --- |
| 01 | Simple | Calculate the Sum of Stair Numbers from 1 to n |
| 02 | Simple | Calculate the Total Fixed Coin Reward Over Several Days |
| 03 | Simple | Calculate the Total Distance Run Over Several Trainings |
| 04 | Normal | Calculate the Sum of All Even Numbers from 1 to n |
| 05 | Normal | Count Reward Levels and Calculate Total Reward Coins |
| 06 | Normal | Calculate Quiz Total, Average, and Pass Status |
| 07 | Challenge | Find the Earliest Day a Growing Daily Savings Plan Reaches the Target |
| 08 | Challenge | Accumulate Coins Through Levels and Return the Level That Reaches the Upgrade Target |
| 09 | Challenge | Count, Sum, and Average Numbers That Meet Conditions |

### Progression Notes

| Difficulty | Combination Method |
| --- | --- |
| Simple | Practices only `sum = sum + value`. |
| Normal | Combines conditions, counting, and averages. |
| Challenge | Combines accumulation, filtering, target checks, and loop stopping. |

---

## Week 05: Factors, Multiples, and Prime Numbers

| Item | Content |
| --- | --- |
| Core theme | Use enumeration to solve number-property problems |
| Math foundation | Factors, multiples, prime numbers, common factors, common multiples |
| Programming content | `%`, loop enumeration, Boolean variables |
| Algorithm content | Factor enumeration, factor counting, prime checks, greatest common factor enumeration, least common multiple enumeration |
| Scenarios | Group competitions, puzzle blocks, special numbers, running groups, alarm cycles |
| Goal | Students understand the algorithmic idea of trying candidates one by one. |

### Teaching Task Design

| No. | Difficulty | Title |
| --: | --- | --- |
| 01 | Simple | Check Whether One Number Is a Multiple of Another |
| 02 | Simple | Check Whether a Team Can Be Split into Equal Groups |
| 03 | Simple | Count Multiples of a Target Number from 1 to n |
| 04 | Normal | Find How Many Equal-Sharing Methods a Puzzle Block Count Has |
| 05 | Normal | Count How Many Factors a Special Number Has |
| 06 | Normal | Calculate the Count and Sum of All Factors of a Number |
| 07 | Challenge | Check Whether a Special Number Is Prime |
| 08 | Challenge | Find the Largest Common Group Size for Two Classes |
| 09 | Challenge | Find the First Minute When Two Alarms Ring Together |

### Progression Notes

| Difficulty | Combination Method |
| --- | --- |
| Simple | Practices multiple and divisibility checks. |
| Normal | Combines loops, counting, and accumulation. |
| Challenge | Combines factor enumeration, prime checks, greatest common factors, and least common multiples. |

---

## Week 06: Quantity Relations and Double Enumeration

| Item | Content |
| --- | --- |
| Core theme | How to find answers when two variables change together |
| Math foundation | Sum and difference relations, multiplication relations, chicken-and-rabbit problems, shopping plans |
| Programming content | Nested loops, condition filtering |
| Algorithm content | Double-variable enumeration, combination counting, plan search, best-plan update |
| Scenarios | Candy sharing, rectangle puzzles, dice games, stationery purchases, chicken-and-rabbit problems |
| Goal | Students can use nested loops to solve simple applied problems. |

### Teaching Task Design

| No. | Difficulty | Title |
| --: | --- | --- |
| 01 | Simple | Count Ways to Buy total Candies from Two Candy Types |
| 02 | Simple | Count Outcomes Where Two Dice Sum to the Target |
| 03 | Simple | Count Length-Width Pairs for a Fixed Rectangle Perimeter |
| 04 | Normal | Count Stationery Plans That Spend a Fixed Budget on Pencils and Erasers |
| 05 | Normal | Count Game Shop Plans That Spend All Coins on Potions and Shields |
| 06 | Normal | Find the Length and Width with Maximum Area for a Fixed Perimeter |
| 07 | Challenge | Chicken and Rabbit Cage: Find Animal Counts from Heads and Legs |
| 08 | Challenge | Spend the Exact Budget on Children's and Adult Tickets |
| 09 | Challenge | Buy Two Products Within Budget and Maximize Quantity with Minimum Leftover |

### Progression Notes

| Difficulty | Combination Method |
| --- | --- |
| Simple | Practices nested loops and plan counting. |
| Normal | Combines price calculation, area calculation, and maximum updates. |
| Challenge | Combines double-variable enumeration, validity checks, and best-plan updates. |

---

## Week 07: Arrays and Table Data

| Item | Content |
| --- | --- |
| Core theme | How to store and process a group of data |
| Math foundation | Table data, sequence numbers, positions |
| Programming content | Array tables, indexes, traversal, modification |
| Algorithm content | Array access, array modification, array traversal, array copying, accumulated arrays |
| Scenarios | Score sheets, fitness records, daily points, line reversal, ball-passing games |
| Goal | Students understand that an array is a row of numbered boxes. |

### Teaching Task Design

| No. | Difficulty | Title |
| --: | --- | --- |
| 01 | Simple | Read the Score of a Student at a Given Position |
| 02 | Simple | Modify the Points for One Day in a Daily Points Table |
| 03 | Simple | Swap the Data at Two Positions in a Student Line |
| 04 | Normal | Traverse a Score Sheet and Count Passing Students |
| 05 | Normal | Copy a Class Score Sheet as a Backup |
| 06 | Normal | Add an Event Bonus to Every Item in a Daily Coin Table |
| 07 | Challenge | Generate a Daily Accumulated Points Table for Upgrade Progress |
| 08 | Challenge | Generate the New Line Order After a Team Turns Around |
| 09 | Challenge | Record Who Has the Ball After Each Round in a Passing Game |

### Progression Notes

| Difficulty | Combination Method |
| --- | --- |
| Simple | Practices array reading, modification, and swapping. |
| Normal | Combines loops, conditions, and variable updates. |
| Challenge | Combines accumulated arrays, reverse access, state simulation, and loop wraparound. |

---

## Week 08: Array Statistics and Extremes

| Item | Content |
| --- | --- |
| Core theme | Analyze a group of data statistically |
| Math foundation | Total score, average score, highest score, lowest score, range statistics |
| Programming content | Array traversal + condition + variable update |
| Algorithm content | Array summation, array counting, maximum value, minimum value, maximum-value position, second-largest value |
| Scenarios | Class scores, long jump competitions, score ranges, runner-up scores |
| Goal | Students master the pattern of assuming the first item is the answer, then comparing one by one. |

### Teaching Task Design

| No. | Difficulty | Title |
| --: | --- | --- |
| 01 | Simple | Calculate the Total Score of a Class Score Sheet |
| 02 | Simple | Count How Many Scores in a Score Sheet Are Passing |
| 03 | Simple | Calculate the Average Reward in a Daily Coin Array |
| 04 | Normal | Find the Best Athlete in a Long Jump Competition |
| 05 | Normal | Find the Lowest Score and Its Position in a Score Sheet |
| 06 | Normal | Count Game Scores Above a Target Line and Sum Them |
| 07 | Challenge | Find the First and Second Place Scores in a Ranking |
| 08 | Challenge | Find the Difference Between First and Second Place in Long Jump |
| 09 | Challenge | Count Score Ranges and Find the Highest Score Position |

### Progression Notes

| Difficulty | Combination Method |
| --- | --- |
| Simple | Practices array summation, array counting, and array averages. |
| Normal | Combines maximum values, minimum values, and position recording. |
| Challenge | Combines second-largest values, range statistics, position updates, and difference calculation. |

---

## Week 09: Search and Binary Search

| Item | Content |
| --- | --- |
| Core theme | Whether a target exists, where it is, and how to find it faster |
| Math foundation | Number lines, intervals, comparisons, range narrowing |
| Programming content | Linear search, `found`, `position`, `left/right/mid` |
| Algorithm content | Existence search, position search, first-position search, last-position search, binary search |
| Scenarios | Number cards, perfect-score search, retake lists, book numbers, search-efficiency comparison |
| Goal | Students understand the difference between linear search and binary search. |

### Teaching Task Design

| No. | Difficulty | Title |
| --: | --- | --- |
| 01 | Simple | Search a Number-Card Array for a Target Student |
| 02 | Simple | Search a Score Sheet for a Perfect Score |
| 03 | Simple | Find the Position of a Target Item Count in Backpack Slots |
| 04 | Normal | Find the First Failing Student in a Retake List |
| 05 | Normal | Find the Last Perfect-Score Position in Competition Results |
| 06 | Normal | Search for a Target Score and Count How Many Checks Were Made |
| 07 | Challenge | Use Binary Search to Quickly Find a Target Book in Sorted Book Numbers |
| 08 | Challenge | Compare How Many Checks Linear Search and Binary Search Need |
| 09 | Challenge | Search a Sorted Score Table for a Target Score or Insertion Position |

### Progression Notes

| Difficulty | Combination Method |
| --- | --- |
| Simple | Practices linear search and position return. |
| Normal | Combines array traversal, conditions, and search-count statistics. |
| Challenge | Combines sorted arrays, binary search, range narrowing, and insertion-position checks. |

---

## Week 10: Sorting Algorithms I

| Item | Content |
| --- | --- |
| Core theme | Organize a group of data with comparison and swapping |
| Math foundation | Comparison, ascending order, descending order, lining up |
| Programming content | Swapping array elements, nested loops |
| Algorithm content | Selection sort, bubble sort, swap-count statistics |
| Scenarios | Seat changes, height lines, score leaderboards, sorting workload |
| Goal | Students understand the three sorting actions: compare, record a position, and swap. |

### Teaching Task Design

| No. | Difficulty | Title |
| --: | --- | --- |
| 01 | Simple | Swap Two Adjacent Height Values That Are Out of Order |
| 02 | Simple | Find the Position of the Minimum Value in a Score List |
| 03 | Simple | Complete the First Round of Selection Sort |
| 04 | Normal | Use Selection Sort to Organize a Game Score Ranking |
| 05 | Normal | Use Selection Sort to Sort Book Numbers from Small to Large |
| 06 | Normal | Complete One Left-to-Right Pass of Bubble Sort |
| 07 | Challenge | Use Bubble Sort to Arrange Scores from High to Low |
| 08 | Challenge | Count Swaps When Bubble Sorting a Seat Table |
| 09 | Challenge | Compare Selection Sort and Bubble Sort Swap Counts on the Same Data |

### Progression Notes

| Difficulty | Combination Method |
| --- | --- |
| Simple | Practices comparison, position recording, and swapping. |
| Normal | Fully implements selection sort or part of bubble sort. |
| Challenge | Combines full sorting, swap-count statistics, array copying, and algorithm comparison. |

---

## Week 11: Sorting Algorithms II and Applications

| Item | Content |
| --- | --- |
| Core theme | The same sorting problem can be solved with different methods |
| Math foundation | Ranks, middle values, duplicate data, merged data |
| Programming content | Insertion sort, processing after sorting |
| Algorithm content | Insertion sort, middle values, deduplication statistics, merging sorted arrays, tied ranks |
| Scenarios | Line insertion, playing-card organization, top three, merging two class score sheets, tied ranks |
| Goal | Students can distinguish the ideas behind selection sort, bubble sort, and insertion sort. |

### Teaching Task Design

| No. | Difficulty | Title |
| --: | --- | --- |
| 01 | Simple | Insert a New Score into the Correct Position Like Organizing Playing Cards |
| 02 | Simple | Find the Middle Score in a Sorted Score Sheet |
| 03 | Simple | Count Adjacent Equal Scores After Sorting |
| 04 | Normal | Use Insertion Sort to Organize Game Scores |
| 05 | Normal | Merge Two Already Sorted Score Sheets |
| 06 | Normal | Sort Scores and Count How Many Different Scores There Are |
| 07 | Challenge | Handle Tied Ranks in a Score Ranking |
| 08 | Challenge | Merge Two Sorted Book Number Lists, Remove Duplicates, and Keep Order |
| 09 | Challenge | Find the Top Three After Sorting and Check Whether a New Score Enters the Top Three |

### Progression Notes

| Difficulty | Combination Method |
| --- | --- |
| Simple | Practices insertion ideas, positions after sorting, and adjacent comparison. |
| Normal | Combines full insertion sort, merging sorted arrays, and deduplication statistics. |
| Challenge | Combines sorting, deduplication, merging, tied ranking, and top-three checks. |

---

## Week 12: Integrated Math Algorithm Projects

| Item | Content |
| --- | --- |
| Core theme | Put algorithms back into real math scenarios |
| Math foundation | Statistics, sorting, search, quantity relations, geometric area |
| Programming content | Integrated use of variables, conditions, loops, arrays, and functions |
| Algorithm content | Statistics, search, sorting, ranking, enumeration, best-plan search |
| Scenarios | Score systems, leaderboards, shopping plans, dice possibilities, book numbers, sports day results |
| Goal | Students can complete a small integrated algorithm project. |

### Teaching Task Design

| No. | Difficulty | Title |
| --: | --- | --- |
| 01 | Simple | Class Score Analysis System: Total, Average, and Passing Count |
| 02 | Simple | Game Points Analysis System: Highest, Lowest, and Average |
| 03 | Simple | Daily Savings Record System: Total, Highest Day, and Goal Check |
| 04 | Normal | Game Leaderboard System: Sorting, Top Three, and Entry Check |
| 05 | Normal | Book Number Management System: Search, Sorting, and Classification |
| 06 | Normal | Sports Day Results System: Best Result, Ranking, and Range Statistics |
| 07 | Challenge | Shopping Plan System: Find the Most Items Within Budget with the Least Leftover |
| 08 | Challenge | Dice Possibility System: Count All Sums and Find the Most Frequent Result |
| 09 | Challenge | Integrated Score Ranking System: Sorting, Tied Ranking, and Student Rank Search |

### Progression Notes

| Difficulty | Combination Method |
| --- | --- |
| Simple | Each question reviews 2 to 3 foundational algorithms. |
| Normal | Combines arrays, search, sorting, and classification statistics. |
| Challenge | Combines enumeration, array statistics, sorting, tied ranking, search, and project modeling. |

---

---

# 10. Week 12 Project Directions

## Project A: Score Analysis System

Fixed data example:

```lua
local scores = {80, 95, 67, 100, 58, 72}
local n = 6
```

Functions:

```text
Calculate total score
Calculate average score
Find highest score
Find lowest score
Count passing students
Count excellent students
Search whether there is a perfect score
Sort from high to low
Calculate the top three
Handle tied ranks
```

---

## Project B: Game Leaderboard System

Fixed data example:

```lua
local points = {120, 300, 250, 180, 400}
local n = 5
```

Functions:

```text
Find highest score
Find lowest score
Count players above a score line
Sort
Find the top three
Check whether a score can enter the top three
Find the rank of a score
```

---

## Project C: Shopping Plan System

Fixed data example:

```lua
local priceA = 3
local priceB = 5
local budget = 30
```

Functions:

```text
Count plans that spend the budget exactly
Count plans that do not exceed the budget
Find the plan with the largest item count
Find the plan with the least leftover money
```

---

## Project D: Dice Possibility Statistics

Fixed data example:

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

---

## Project E: Book Number Management System

Fixed data example:

```lua
local books = {18, 5, 12, 3}
local n = 4
```

Functions:

```text
Search for a target number
Count odd-numbered IDs
Count even-numbered IDs
Sort from small to large
Merge two sorted ID lists
Remove duplicate IDs
```

---

# 11. Core Algorithm Coverage Checklist

| No. | Algorithm |
| --: | --- |
| 01 | Sequential calculation |
| 02 | Variable update |
| 03 | Variable swap |
| 04 | Conditional classification |
| 05 | Odd-even check |
| 06 | Multiple check |
| 07 | Loop counting |
| 08 | Multiple-condition check |
| 09 | Conditional counting |
| 10 | Classification statistics |
| 11 | Accumulation algorithm |
| 12 | Conditional summation |
| 13 | Average |
| 14 | Target accumulation |
| 15 | Factor enumeration |
| 16 | Factor counting |
| 17 | Prime check |
| 18 | Greatest common factor enumeration |
| 19 | Least common multiple enumeration |
| 20 | Single-variable enumeration |
| 21 | Double-variable enumeration |
| 22 | Plan counting |
| 23 | Best-plan update |
| 24 | Array access |
| 25 | Array modification |
| 26 | Array swap |
| 27 | Array traversal |
| 28 | Array copy |
| 29 | Array reversal |
| 30 | Accumulated array |
| 31 | Array summation |
| 32 | Array counting |
| 33 | Maximum / minimum value |
| 34 | Maximum-value position |
| 35 | Second-largest value |
| 36 | Linear search |
| 37 | Find first position |
| 38 | Find last position |
| 39 | Search-count statistics |
| 40 | Binary search |
| 41 | Insertion-position check |
| 42 | Selection sort |
| 43 | Bubble sort |
| 44 | Insertion sort |
| 45 | Find top three after sorting |
| 46 | Deduplicate after sorting |
| 47 | Merge sorted arrays |
| 48 | Tied ranking |
| 49 | Linked array sorting |
| 50 | Integrated project modeling |

---

# 12. Corresponding Elementary Math Content

| Math Area | Course Content |
| --- | --- |
| Numbers and operations | Arithmetic, multiples, factors, prime numbers, remainders, averages |
| Quantity relations | Candy sharing, shopping, chicken-and-rabbit problems, combination enumeration, best plans |
| Statistics and probability | Score statistics, range statistics, dice possibilities, frequencies |
| Shapes and geometry | Rectangle area, perimeter, area comparison |
| Integration and practice | Score systems, leaderboards, shopping plans, dice analysis, book number management |

---

# 13. Lesson and Question Naming Rules

## Lesson Naming Rules

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

---

## Question Field Rules

Each question uses only these index fields:

```text
No.
Difficulty
Title
```

Question details continue to use these fields:

```text
Story / Background
Task
Input
Output
Hints
Starter Code
Reference Solution
Test Cases
```

Each lesson has a fixed structure:

```text
Simple 01-03
Normal 04-06
Challenge 07-09
```

---

# 14. Standard Template for Future Lesson Generation

Each lesson is recommended to contain these fields:

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

Each lesson is recommended to contain these sections:

```text
1. Lesson objectives
2. Lesson limits
3. Core concepts
4. 9 scenario-based questions
5. Lesson review
```

---

# 15. Course Progression Summary

```text
Week 01: Build the foundation for variables and sequential calculation
Week 02: Build condition, loop, and state-update skills
Week 03: Move from loops to counting and classification statistics
Week 04: Extend from counting to summation, averages, and target accumulation
Week 05: Use enumeration to handle factors, multiples, and prime numbers
Week 06: Use nested loops to handle problems where two variables change together
Week 07: Introduce arrays and begin processing groups of data
Week 08: Perform statistics and extreme-value analysis on arrays
Week 09: Search for targets in arrays and understand binary search
Week 10: Understand comparison and swapping through selection sort and bubble sort
Week 11: Learn insertion sort, merging, deduplication, and ranking
Week 12: Integrate statistics, search, sorting, enumeration, and modeling through projects
```

Overall progression:

```text
Variables -> Conditions and loops -> Counting -> Accumulation -> Enumeration -> Double enumeration -> Arrays -> Array statistics -> Search -> Sorting -> Sorting applications -> Integrated projects
```
