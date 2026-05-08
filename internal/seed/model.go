package seed

const (
	demoTeacherUsername  = "teacher"
	demoTeacherPassword  = "teacher"
	demoStudentUsername  = "student"
	demoStudentPassword  = "student"
	demoClassroomName    = "teacher_demo_classroom"
	exampleClassroomName = "example_classroom"
	embeddedLessonCount  = 24
)

type lessonSeed struct {
	Title       string
	Description string
	SortOrder   int
	Questions   []lessonQuestionSeed
}

type lessonQuestionSeed struct {
	Title         string
	Description   string
	StarterCode   string
	ReferenceCode string
	TestCases     string
	SortOrder     int
}
