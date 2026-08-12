---
name: generate-course-question-bank
description: Generate complete, age-appropriate programming courses and import-ready lesson question banks for oj-lite. Use when a teacher asks to plan a course, create one or more lessons, write Lua coding exercises, generate lesson JSON, expand an existing curriculum, translate a question bank, or review course content for pedagogical and judge compatibility.
---

# Generate an oj-lite Course Question Bank

Create coherent programming courses that teachers can import into oj-lite. Optimize for clear learning progression, classroom usability, valid lesson JSON, and exercises that run in oj-lite's restricted Lua judge.

## Establish the Course Brief

Determine the following before generating the course:

- Student age or experience level
- Course language
- Learning goals and prerequisites
- Number of lessons and questions per lesson
- Topic sequence
- Difficulty distribution
- First lesson `sort_order`
- Any allowed or forbidden Lua features

Ask concise questions only when missing information would materially change the course. Otherwise, state reasonable assumptions and continue.

Use these defaults when the teacher provides no preference:

- Target students who are moving from visual programming to beginner text programming.
- Use the language of the teacher's request for student-facing content.
- Create 9 questions per lesson: 3 Simple, 3 Normal, and 3 Challenge.
- Give every lesson one focused topic and one measurable learning objective.
- Increase difficulty within each lesson and build cumulatively across lessons.
- Start `sort_order` at 1, but remind the teacher that lesson order must be globally unique in their oj-lite installation.
- Include at least 3 test cases per question: a typical case, a boundary case, and a case likely to expose a common mistake.

## Plan Before Writing Questions

Create a compact course outline containing, for each lesson:

1. Lesson title
2. Learning objective
3. New concepts
4. Prerequisite concepts reviewed
5. Planned Simple, Normal, and Challenge questions

Keep the scope realistic. Introduce only a small number of new ideas per lesson. Make Simple questions isolate the new concept, Normal questions combine it with prior knowledge, and Challenge questions require decomposition without relying on concepts that have not yet been taught.

If the teacher requests a large course, keep the outline internally consistent and generate every requested lesson. Do not silently shorten the course. If output limits prevent completing it at once, divide delivery by lesson boundaries and clearly identify the next lesson to generate.

## Write Import-Ready Lesson JSON

Generate one JSON object per lesson. Do not wrap multiple lessons in a new course-level array or object unless the teacher explicitly requests that format. When working in a repository and asked to save the result, use one `.json` file per lesson. Otherwise, return one fenced `json` block per lesson.

Use this exact structure for a new lesson:

```json
{
  "title": "Week 01 - Variables",
  "description": "Practice sequential calculations and variable updates.",
  "sort_order": 1,
  "questions": [
    {
      "title": "Calculate the Remaining Health",
      "description": "## Difficulty\n\nSimple\n\n## Background\n\nA game character is attacked once.\n\n## Task\n\nCalculate the remaining health.\n\n## Input\n\nTwo numbers: hp and damage.\n\n## Output\n\nReturn the remaining health.\n\n## Hint\n\nUpdate the health with one subtraction.",
      "starter_code": "function solution(hp, damage)\n    return 0\nend",
      "reference_code": "function solution(hp, damage)\n    return hp - damage\nend",
      "test_cases": [
        { "input": [100, 25] },
        { "input": [50, 0] },
        { "input": [20, 20] }
      ],
      "sort_order": 1
    }
  ]
}
```

Follow these schema rules:

- Make the lesson `title` non-empty.
- Make lesson `sort_order` a positive integer that is globally unique.
- Make every question `title` non-empty.
- Make question `sort_order` a positive integer unique within its lesson; normally use consecutive values starting at 1.
- Omit question `id` when creating new questions. Preserve real IDs only when the teacher explicitly provides an exported lesson for replacement.
- Store the question statement as a Markdown string in `description`.
- Emit strict JSON with double-quoted keys and strings, no comments, and no trailing commas.
- Escape newlines inside JSON strings as `\n`.
- Keep each lesson self-contained so it can be pasted into the local admin Lesson JSON editor.

## Design Each Question

Use this Markdown structure inside every question `description`:

```markdown
## Difficulty

Simple | Normal | Challenge

## Background

A short, age-appropriate context.

## Task

An unambiguous action for the student.

## Input

Parameter names, types, meanings, ranges, and important guarantees.

## Output

The exact return value or ordered return values.

## Hint

A direction that does not reveal the full solution.
```

Apply these writing rules:

- Keep the title, statement, function parameters, return values, reference solution, and tests perfectly aligned.
- Specify all constraints needed to avoid undefined behavior, such as non-empty arrays or non-zero divisors.
- Prefer familiar, respectful scenarios without unnecessary story text.
- Avoid cultural assumptions, stereotypes, trick wording, and facts unrelated to the programming goal.
- Make the challenge come from reasoning, not from long calculations or confusing prose.
- Do not reveal the complete condition, loop bounds, formula, or code in the hint.
- Avoid near-duplicate questions that only rename variables or change the story.

## Follow the oj-lite Lua Judge Model

Write both `starter_code` and `reference_code` as source code defining this global function:

```lua
function solution(...)
    -- implementation
end
```

The judge passes each test case as function arguments and compares all returned values from the student solution with all returned values from `reference_code`.

Respect these runtime constraints:

- Use Lua only.
- Do not read input or print the answer; return the answer from `solution`.
- Do not use `io`, `os`, `package`, `debug`, `require`, `dofile`, or `loadfile`.
- Do not use unavailable standard libraries such as `math`, `string`, or `table`.
- Do not rely on `pairs`, `ipairs`, or other standard-library globals.
- Use language primitives such as local variables, arithmetic, comparisons, `if`, `for`, `while`, functions, table literals, and numeric table indexes.
- Keep algorithms comfortably within the 2-second limit for every test case.
- Prefer deterministic integer arithmetic. When teaching division, choose inputs and requirements that avoid ambiguous floating-point answers because return values are compared exactly.
- Return only JSON-compatible values: nil, booleans, numbers, strings, and tables composed of those values.

Make `starter_code` helpful but incomplete. Preserve the exact function name and parameter list from the reference solution. Use neutral placeholder returns of the correct count and broad type where practical, without exposing the algorithm.

## Build Effective Test Cases

`test_cases` contains inputs only. Do not include expected outputs; oj-lite obtains them by executing `reference_code`.

Prefer the explicit form:

```json
"test_cases": [
  { "input": [10, 3] },
  { "input": [0, 5] },
  { "input": [-4, -7] }
]
```

Each item in `input` is one argument to `solution`. For an array argument, nest the array:

```json
{ "input": [[4, 1, 9, 2], 4] }
```

Use JSON numbers, strings, booleans, arrays, objects, or null. Design tests to cover:

- The ordinary path
- Minimum and maximum meaningful values
- Equality and threshold boundaries
- Empty, single-item, repeated-value, sorted, or reverse-sorted arrays when allowed by the statement
- Cases that distinguish the intended algorithm from common incorrect shortcuts

Never add a test that violates the guarantees stated in the question. Keep loops bounded and test values small enough for classroom feedback to remain fast.

## Validate the Course

Perform all checks before delivering the result:

1. Parse every lesson as strict JSON.
2. Confirm each lesson and question `sort_order` is positive and unique in its required scope.
3. Confirm every question has all required fields.
4. Confirm each description's inputs and outputs match the Lua signature and return values.
5. Confirm starter and reference code use the same `solution` parameters.
6. Execute or mentally trace the reference solution against every test case.
7. Confirm the reference solution uses no unavailable libraries or globals.
8. Confirm difficulty and prerequisites progress consistently across the course.
9. Confirm hints guide without giving away the answer.
10. Confirm the final output contains no prose inside JSON blocks.

When filesystem and runtime tools are available, validate JSON with a parser and run representative reference solutions through the project's tests or judge. Report assumptions and any validation that could not be performed.

## Deliver to the Teacher

Provide:

1. A brief summary of the course structure and assumptions
2. The complete import-ready lesson JSON objects or saved file paths
3. A short validation summary
4. A reminder to verify lesson `sort_order` values against existing lessons before import

Do not expose internal chain-of-thought. Keep explanations practical and teacher-facing.
