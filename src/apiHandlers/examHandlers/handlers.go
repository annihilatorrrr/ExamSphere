package examHandlers

import (
	"ExamSphere/src/apiHandlers"
	"ExamSphere/src/core/utils/logging"
	"ExamSphere/src/database"
	"time"

	"github.com/ALiwoto/ssg/ssg"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

// CreateExamV1 godoc
// @Summary Create a new exam
// @Description Allows the user to create a new exam.
// @ID createExamV1
// @Tags Exam
// @Accept json
// @Produce json
// @Param Authorization header string true "Authorization token"
// @Param data body CreateExamData true "Data needed to create a new exam"
// @Success 200 {object} apiHandlers.EndpointResponse{result=CreateExamResult}
// @Router /api/v1/exam/create [post]
func CreateExamV1(c *fiber.Ctx) error {
	claimInfo := apiHandlers.GetJWTClaimsInfo(c)
	if claimInfo == nil {
		return apiHandlers.SendErrInvalidJWT(c)
	}

	userInfo := database.GetUserInfoByAuthHash(
		claimInfo.UserId, claimInfo.AuthHash,
	)

	if userInfo == nil {
		return apiHandlers.SendErrInvalidAuth(c)
	} else if !userInfo.CanCreateNewExam() {
		return apiHandlers.SendErrPermissionDenied(c)
	}

	data := &CreateExamData{}
	if err := c.BodyParser(data); err != nil {
		return apiHandlers.SendErrInvalidBodyData(c)
	}

	if !data.IsValid() {
		return apiHandlers.SendErrInvalidBodyData(c)
	}

	examInfo, err := database.CreateNewExam(&database.NewExamData{
		CourseId:        data.CourseId,
		ExamTitle:       data.ExamTitle,
		ExamDescription: data.ExamDescription,
		Price:           data.Price,
		IsPublic:        data.IsPublic,
		Duration:        data.Duration,
		ExamDate:        time.Unix(data.ExamDate, 0),
		CreatedBy:       userInfo.UserId,
	})

	if err != nil {
		return apiHandlers.SendErrInternalServerError(c)
	}

	return apiHandlers.SendResult(c, &CreateExamResult{
		ExamId:    examInfo.ExamId,
		CourseId:  examInfo.CourseId,
		Price:     examInfo.Price,
		CreatedAt: examInfo.CreatedAt,
		ExamDate:  examInfo.ExamDate,
		Duration:  examInfo.Duration,
		CreatedBy: examInfo.CreatedBy,
		IsPublic:  examInfo.IsPublic,
	})
}

// GetExamInfoV1 godoc
// @Summary Get information about an exam
// @Description Allows the user to get information about an exam.
// @ID getExamInfoV1
// @Tags Exam
// @Accept json
// @Produce json
// @Param Authorization header string true "Authorization token"
// @Param id query int true "Exam ID"
// @Success 200 {object} apiHandlers.EndpointResponse{result=GetExamInfoResult}
// @Router /api/v1/exam/info [get]
func GetExamInfoV1(c *fiber.Ctx) error {
	claimInfo := apiHandlers.GetJWTClaimsInfo(c)
	if claimInfo == nil {
		return apiHandlers.SendErrInvalidJWT(c)
	}

	userInfo := database.GetUserInfoByAuthHash(
		claimInfo.UserId, claimInfo.AuthHash,
	)

	if userInfo == nil {
		return apiHandlers.SendErrInvalidAuth(c)
	} else if !userInfo.CanGetExamInfo() {
		return apiHandlers.SendErrPermissionDenied(c)
	}

	examId := c.QueryInt("id")
	if examId == 0 {
		return apiHandlers.SendErrParameterRequired(c, "id")
	}

	examInfo, err := database.GetExamInfo(examId)
	if err == database.ErrExamNotFound || examInfo == nil {
		return apiHandlers.SendErrExamNotFound(c)
	} else if err != nil {
		return apiHandlers.SendErrInternalServerError(c)
	}

	return apiHandlers.SendResult(c, &GetExamInfoResult{
		ExamId:          examInfo.ExamId,
		CourseId:        examInfo.CourseId,
		ExamTitle:       examInfo.ExamTitle,
		ExamDescription: examInfo.ExamDescription,
		Price:           examInfo.Price,
		CreatedAt:       examInfo.CreatedAt,
		ExamDate:        examInfo.ExamDate,
		Duration:        examInfo.Duration,
		CreatedBy:       examInfo.CreatedBy,
		IsPublic:        examInfo.IsPublic,
		HasStarted:      examInfo.HasExamStarted(),
		HasFinished:     examInfo.HasExamFinished(),
		StartsIn:        examInfo.ExamStartsIn(),
		FinishesIn:      examInfo.ExamFinishesIn(),
		QuestionCount:   database.GetExamQuestionsCount(examId),
	})
}

// EditExamV1 godoc
// @Summary Edit an exam
// @Description Allows the user to edit an exam.
// @ID editExamV1
// @Tags Exam
// @Accept json
// @Produce json
// @Param Authorization header string true "Authorization token"
// @Param data body EditExamData true "Data needed to edit an exam"
// @Success 200 {object} apiHandlers.EndpointResponse{result=EditExamResult}
// @Router /api/v1/exam/edit [post]
func EditExamV1(c *fiber.Ctx) error {
	claimInfo := apiHandlers.GetJWTClaimsInfo(c)
	if claimInfo == nil {
		return apiHandlers.SendErrInvalidJWT(c)
	}

	userInfo := database.GetUserInfoByAuthHash(
		claimInfo.UserId, claimInfo.AuthHash,
	)

	if userInfo == nil {
		return apiHandlers.SendErrInvalidAuth(c)
	} else if !userInfo.CanTryToEditExam() {
		return apiHandlers.SendErrPermissionDenied(c)
	}

	data := &EditExamData{}
	if err := c.BodyParser(data); err != nil {
		return apiHandlers.SendErrInvalidBodyData(c)
	}

	if !data.IsValid() {
		return apiHandlers.SendErrInvalidBodyData(c)
	}

	examInfo := database.GetExamInfoOrNil(data.ExamId)
	if examInfo == nil {
		return apiHandlers.SendErrExamNotFound(c)
	}

	if !userInfo.CanEditExam(examInfo) {
		return apiHandlers.SendErrPermissionDenied(c)
	}

	examInfo, err := database.EditExamInfo(&database.EditExamInfoData{
		ExamId:          data.ExamId,
		CourseId:        data.CourseId,
		ExamTitle:       data.ExamTitle,
		ExamDescription: data.ExamDescription,
		Price:           data.Price,
		IsPublic:        data.IsPublic,
		Duration:        data.Duration,
		ExamDate:        time.Unix(data.ExamDate, 0),
	})

	if err != nil {
		return apiHandlers.SendErrInternalServerError(c)
	}

	return apiHandlers.SendResult(c, &EditExamResult{
		ExamId:          examInfo.ExamId,
		CourseId:        examInfo.CourseId,
		ExamTitle:       examInfo.ExamTitle,
		ExamDescription: examInfo.ExamDescription,
		Price:           examInfo.Price,
		CreatedAt:       examInfo.CreatedAt,
		ExamDate:        examInfo.ExamDate,
		Duration:        examInfo.Duration,
		CreatedBy:       examInfo.CreatedBy,
		IsPublic:        examInfo.IsPublic,
	})
}

// GetExamQuestionsV1 godoc
// @Summary Get questions of an exam
// @Description Allows the user to get questions of an exam.
// @ID getExamQuestionsV1
// @Tags Exam
// @Accept json
// @Produce json
// @Param Authorization header string true "Authorization token"
// @Param data body GetExamQuestionsData true "Data needed to get questions of an exam"
// @Success 200 {object} apiHandlers.EndpointResponse{result=GetExamQuestionsResult}
// @Router /api/v1/exam/questions [post]
func GetExamQuestionsV1(c *fiber.Ctx) error {
	claimInfo := apiHandlers.GetJWTClaimsInfo(c)
	if claimInfo == nil {
		return apiHandlers.SendErrInvalidJWT(c)
	}

	userInfo := database.GetUserInfoByAuthHash(
		claimInfo.UserId, claimInfo.AuthHash,
	)

	if userInfo == nil {
		return apiHandlers.SendErrInvalidAuth(c)
	} else if !userInfo.CanGetExamQuestions() {
		return apiHandlers.SendErrPermissionDenied(c)
	}

	data := &GetExamQuestionsData{}
	if err := c.BodyParser(data); err != nil {
		return apiHandlers.SendErrInvalidBodyData(c)
	}

	if data.ExamId == 0 {
		return apiHandlers.SendErrParameterRequired(c, "exam_id")
	}

	examInfo := database.GetExamInfoOrNil(data.ExamId)
	if examInfo == nil {
		return apiHandlers.SendErrExamNotFound(c)
	}

	if !userInfo.CanPeekExamQuestions(examInfo.CreatedBy) &&
		(!database.HasParticipatedInExam(userInfo.UserId, data.ExamId) ||
			!examInfo.HasExamStarted()) {
		return apiHandlers.SendErrNotParticipatedInExam(c)
	}

	questions, err := database.GetExamQuestions(data.ExamId)
	if err != nil && err != pgx.ErrNoRows {
		return apiHandlers.SendErrInternalServerError(c)
	}

	questionsInfo := make([]*ExamQuestionInfo, 0, len(questions))
	for _, q := range questions {
		info := &ExamQuestionInfo{
			QuestionId:    q.QuestionId,
			QuestionTitle: q.QuestionTitle,
			Description:   q.Description,
			Option1:       q.Option1,
			Option2:       q.Option2,
			Option3:       q.Option3,
			Option4:       q.Option4,
			CreatedAt:     q.CreatedAt,
		}

		givenAnswer := database.GetGivenAnswerOrNil(&database.GetGivenAnswerData{
			ExamId:     q.ExamId,
			QuestionId: q.QuestionId,
			UserId:     userInfo.UserId,
		})
		if givenAnswer != nil {
			info.UserAnswer = &AnsweredQuestionInfo{
				QuestionId:   q.QuestionId,
				ChosenOption: ssg.Clone(givenAnswer.ChosenOption),
				SecondsTaken: givenAnswer.SecondsTaken,
				AnswerText:   ssg.Clone(givenAnswer.AnswerText),
			}
		}
		questionsInfo = append(questionsInfo, info)
	}

	return apiHandlers.SendResult(c, &GetExamQuestionsResult{
		ExamId:    data.ExamId,
		Questions: questionsInfo,
	})
}

// AnswerExamQuestionV1 godoc
// @Summary Answer a question of an exam
// @Description Allows the user to answer a question of an exam.
// @ID answerExamQuestionV1
// @Tags Exam
// @Accept json
// @Produce json
// @Param Authorization header string true "Authorization token"
// @Param data body AnswerQuestionData true "Data needed to answer a question of an exam"
// @Success 200 {object} apiHandlers.EndpointResponse{result=AnswerQuestionResult}
// @Router /api/v1/exam/answer [post]
func AnswerExamQuestionV1(c *fiber.Ctx) error {
	claimInfo := apiHandlers.GetJWTClaimsInfo(c)
	if claimInfo == nil {
		return apiHandlers.SendErrInvalidJWT(c)
	}

	userInfo := database.GetUserInfoByAuthHash(
		claimInfo.UserId, claimInfo.AuthHash,
	)

	if userInfo == nil {
		return apiHandlers.SendErrInvalidAuth(c)
	}

	data := &AnswerQuestionData{}
	if err := c.BodyParser(data); err != nil {
		return apiHandlers.SendErrInvalidBodyData(c)
	}

	if data.ExamId == 0 {
		return apiHandlers.SendErrParameterRequired(c, "exam_id")
	} else if data.QuestionId == 0 {
		return apiHandlers.SendErrParameterRequired(c, "question_id")
	} else if data.ChosenOption == nil && data.AnswerText == nil {
		return apiHandlers.SendErrParameterRequired(c, "chosen_option or answer")
	}

	examInfo := database.GetExamInfoOrNil(data.ExamId)
	if examInfo == nil {
		return apiHandlers.SendErrExamNotFound(c)
	}

	if !examInfo.HasExamStarted() {
		return apiHandlers.SendErrExamNotStarted(c)
	} else if examInfo.HasExamFinished() {
		return apiHandlers.SendErrExamFinished(c)
	}

	question, err := database.GetExamQuestion(data.ExamId, data.QuestionId)
	if err != nil {
		if err == database.ErrExamQuestionNotFound {
			return apiHandlers.SendErrExamQuestionNotFound(c)
		}
		logging.UnexpectedError("GetExamQuestion: Failed to get exam question info:", err)
		return apiHandlers.SendErrInternalServerError(c)
	} else if question == nil {
		return apiHandlers.SendErrInternalServerError(c)
	}

	if !database.HasParticipatedInExam(userInfo.UserId, data.ExamId) {
		return apiHandlers.SendErrNotParticipatedInExam(c)
	}

	if data.ChosenOption != nil && !question.HasOption(*data.ChosenOption) {
		return apiHandlers.SendErrInvalidAnswerOption(c)
	}

	givenAnswer, err := database.AnswerQuestion(&database.AnswerQuestionData{
		ExamId:       data.ExamId,
		QuestionId:   data.QuestionId,
		AnsweredBy:   userInfo.UserId,
		ChosenOption: data.ChosenOption,
		SecondsTaken: data.SecondsTaken,
		AnswerText:   data.AnswerText,
	})
	if err != nil {
		logging.UnexpectedError("AnswerQuestion: Failed to answer question:", err)
		return apiHandlers.SendErrInternalServerError(c)
	} else if givenAnswer == nil {
		logging.UnexpectedError("AnswerQuestion: database returned nil for givenAnswer, with no errors")
		return apiHandlers.SendErrInternalServerError(c)
	}

	return apiHandlers.SendResult(c, &AnswerQuestionResult{
		ExamId:     givenAnswer.ExamId,
		QuestionId: givenAnswer.QuestionId,
		AnsweredBy: givenAnswer.AnsweredBy,
		AnsweredAt: givenAnswer.AnsweredAt,
	})
}

// SetScoreV1 godoc
// @Summary Set score for a user in an exam
// @Description Allows the user to set score for a user in an exam.
// @ID setScoreV1
// @Tags Exam
// @Accept json
// @Produce json
// @Param Authorization header string true "Authorization token"
// @Param data body SetScoreData true "Data needed to set score for a user in an exam"
// @Success 200 {object} apiHandlers.EndpointResponse{result=SetScoreResult}
// @Router /api/v1/exam/setScore [post]
func SetScoreV1(c *fiber.Ctx) error {
	claimInfo := apiHandlers.GetJWTClaimsInfo(c)
	if claimInfo == nil {
		return apiHandlers.SendErrInvalidJWT(c)
	}

	userInfo := database.GetUserInfoByAuthHash(
		claimInfo.UserId, claimInfo.AuthHash,
	)

	if userInfo == nil {
		return apiHandlers.SendErrInvalidAuth(c)
	} else if !userInfo.CanTryToScoreExam() {
		return apiHandlers.SendErrPermissionDenied(c)
	}

	data := &SetScoreData{}
	if err := c.BodyParser(data); err != nil {
		return apiHandlers.SendErrInvalidBodyData(c)
	}

	if data.ExamId == 0 {
		return apiHandlers.SendErrParameterRequired(c, "exam_id")
	} else if data.UserId == "" {
		return apiHandlers.SendErrParameterRequired(c, "user_id")
	} else if data.Score == "" {
		return apiHandlers.SendErrParameterRequired(c, "score")
	}

	examInfo := database.GetExamInfoOrNil(data.ExamId)
	if examInfo == nil {
		return apiHandlers.SendErrExamNotFound(c)
	}

	if examInfo.CreatedBy != userInfo.UserId &&
		!userInfo.CanForceScoreExam() {
		return apiHandlers.SendErrPermissionDenied(c)
	}

	if !database.HasParticipatedInExam(data.UserId, data.ExamId) {
		return apiHandlers.SendErrNotParticipatedInExam(c)
	}

	scoreInfo, err := database.SetScoreForUserInExam(&database.NewScoreData{
		ExamId:     data.ExamId,
		UserId:     data.UserId,
		FinalScore: data.Score,
		ScoredBy:   userInfo.UserId,
	})

	if err != nil {
		logging.UnexpectedError("SetExamScore: Failed to set exam score:", err)
		return apiHandlers.SendErrInternalServerError(c)
	} else if scoreInfo == nil {
		logging.UnexpectedError("SetExamScore: database returned nil for scoreInfo, with no errors")
		return apiHandlers.SendErrInternalServerError(c)
	}

	return apiHandlers.SendResult(c, &SetScoreResult{
		ExamId:   scoreInfo.ExamId,
		UserId:   scoreInfo.UserId,
		Score:    *scoreInfo.FinalScore,
		ScoredBy: *scoreInfo.ScoredBy,
	})
}

// GetGivenExamV1 godoc
// @Summary Get information about an exam that a user has participated in
// @Description Allows the user to get information about an exam that a user has participated in.
// @ID getGivenExamV1
// @Tags Exam
// @Accept json
// @Produce json
// @Param Authorization header string true "Authorization token"
// @Param data body GetGivenExamData true "Data needed to get information about an exam that a user has participated in"
// @Success 200 {object} apiHandlers.EndpointResponse{result=GetGivenExamResult}
// @Router /api/v1/exam/givenExam [post]
func GetGivenExamV1(c *fiber.Ctx) error {
	claimInfo := apiHandlers.GetJWTClaimsInfo(c)
	if claimInfo == nil {
		return apiHandlers.SendErrInvalidJWT(c)
	}

	userInfo := database.GetUserInfoByAuthHash(
		claimInfo.UserId, claimInfo.AuthHash,
	)
	if userInfo == nil {
		return apiHandlers.SendErrInvalidAuth(c)
	}

	data := &GetGivenExamData{}
	if err := c.BodyParser(data); err != nil {
		return apiHandlers.SendErrInvalidBodyData(c)
	}

	if data.UserId == "" {
		return apiHandlers.SendErrParameterRequired(c, "user_id")
	} else if data.ExamId == 0 {
		return apiHandlers.SendErrParameterRequired(c, "exam_id")
	}

	examInfo, err := database.GetGivenExam(data.UserId, data.ExamId)
	if err == database.ErrGivenExamNotFound || examInfo == nil {
		return apiHandlers.SendErrGivenExamNotFound(c)
	} else if err != nil {
		logging.UnexpectedError("GetGivenExam: Failed to get given exam info:", err)
		return apiHandlers.SendErrInternalServerError(c)
	}

	return apiHandlers.SendResult(c, &GetGivenExamResult{
		UserId:     examInfo.UserId,
		ExamId:     examInfo.ExamId,
		Price:      examInfo.Price,
		AddedBy:    ssg.Clone(examInfo.AddedBy),
		ScoredBy:   ssg.Clone(examInfo.ScoredBy),
		CreatedAt:  examInfo.CreatedAt,
		FinalScore: ssg.Clone(examInfo.FinalScore),
	})
}

// GetUserOngoingExamsV1 godoc
// @Summary Get ongoing exams of a user
// @Description Allows the user to get ongoing exams of a user.
// @ID getUserOngoingExamsV1
// @Tags Exam
// @Accept json
// @Produce json
// @Param Authorization header string true "Authorization token"
// @Param targetId query string false "Target user id"
// @Success 200 {object} apiHandlers.EndpointResponse{result=GetUserOngoingExamsResult}
// @Router /api/v1/exam/userOngoingExams [get]
func GetUserOngoingExamsV1(c *fiber.Ctx) error {
	claimInfo := apiHandlers.GetJWTClaimsInfo(c)
	if claimInfo == nil {
		return apiHandlers.SendErrInvalidJWT(c)
	}

	userInfo := database.GetUserInfoByAuthHash(
		claimInfo.UserId, claimInfo.AuthHash,
	)
	if userInfo == nil {
		return apiHandlers.SendErrInvalidAuth(c)
	}

	// optional: provide another user's id to see their ongoing exams
	targetUserId := c.Query("targetId")
	if targetUserId == "" {
		targetUserId = userInfo.UserId
	}

	exams, err := database.GetUserOngoingExams(targetUserId)
	if err != nil {
		return apiHandlers.SendErrInternalServerError(c)
	}

	examsInfo := make([]*UserOngoingExamInfo, 0, len(exams))
	for _, exam := range exams {
		examsInfo = append(examsInfo, &UserOngoingExamInfo{
			ExamId:    exam.ExamId,
			ExamTitle: exam.ExamTitle,
			StartTime: exam.StartTime,
		})
	}

	return apiHandlers.SendResult(c, &GetUserOngoingExamsResult{
		Exams: examsInfo,
	})
}

// GetUserExamsHistoryV1 godoc
// @Summary Get history of exams of a user
// @Description Allows the user to get history of exams of a user.
// @ID getUserExamsHistoryV1
// @Tags Exam
// @Accept json
// @Produce json
// @Param Authorization header string true "Authorization token"
// @Param data body GetUsersExamHistoryData true "Data needed to get history of exams of a user"
// @Success 200 {object} apiHandlers.EndpointResponse{result=GetUsersExamHistoryResult}
// @Router /api/v1/exam/userExamsHistory [post]
func GetUserExamsHistoryV1(c *fiber.Ctx) error {
	claimInfo := apiHandlers.GetJWTClaimsInfo(c)
	if claimInfo == nil {
		return apiHandlers.SendErrInvalidJWT(c)
	}

	userInfo := database.GetUserInfoByAuthHash(
		claimInfo.UserId, claimInfo.AuthHash,
	)
	if userInfo == nil {
		return apiHandlers.SendErrInvalidAuth(c)
	}

	data := &GetUsersExamHistoryData{}
	if err := c.BodyParser(data); err != nil {
		return apiHandlers.SendErrInvalidBodyData(c)
	}

	if data.UserId == "" {
		return apiHandlers.SendErrParameterRequired(c, "user_id")
	}

	exams, err := database.GetUserExamsHistory(&database.GetUserExamsHistoryOptions{
		UserId: data.UserId,
		Offset: data.Offset,
		Limit:  data.Limit,
	})
	if err != nil {
		return apiHandlers.SendErrInternalServerError(c)
	}

	examsInfo := make([]*UserExamHistoryInfo, 0, len(exams))
	for _, exam := range exams {
		examsInfo = append(examsInfo, &UserExamHistoryInfo{
			ExamId:    exam.ExamId,
			ExamTitle: exam.ExamTitle,
			StartedAt: exam.StartedAt,
		})
	}

	return apiHandlers.SendResult(c, &GetUsersExamHistoryResult{
		Exams: examsInfo,
	})
}
