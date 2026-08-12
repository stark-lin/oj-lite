(() => {
  'use strict';

  const {
    buildLineNumbers,
    buildJudgeResultMeta,
    bindResizablePane,
    collectElementsById,
    createApiRequest,
    deepClone: clone,
    escapeHtml,
    findById,
    highlightLua,
    normalizeJudgeReport,
    normalizeStringValue,
    normalizeStdoutBuffer,
    redirectToLogin,
    renderJudgeReportPlaceholder,
    renderJudgeReportTerminal,
    renderMarkdown,
    ui
  } = window.OJLite;

  const APP_CONFIG = {
    languageLabel: 'Lua · UTF-8',
    mobileBreakpoint: 900,
    resizer: {
      sidebar: {
        cssVar: '--sidebar-width',
        min: 180,
        max: (rect) => Math.min(420, rect.width * 0.42),
      },
      problem: {
        cssVar: '--problem-width',
        min: 260,
        max: (rect) => Math.min(640, rect.width * 0.62),
      },
      result: {
        cssVar: '--result-height',
        min: 140,
        max: (rect) => Math.min(420, rect.height * 0.55),
      },
    },
    api: {
      me: '/api/me',
      currentLesson: '/api/student/current-lesson',
      questionDetail: (lessonQuestionId) => '/api/student/questions/' + encodeURIComponent(String(lessonQuestionId)),
      questionSubmissions: (lessonQuestionId) => '/api/student/questions/' + encodeURIComponent(String(lessonQuestionId)) + '/submissions',
      submissions: '/api/student/submissions',
      submissionDetail: (submissionId) => '/api/student/submissions/' + encodeURIComponent(String(submissionId)),
    },
  };

  const PLACEHOLDERS = {
    lessonTitleLoading: 'Loading lesson...',
    noLessonTitle: 'No current lesson',
    noQuestions: 'No questions in the current lesson yet.',
    questionDetailLoading: 'Loading question detail...',
    questionDetailUnavailable: 'Question detail is not available yet.',
    submissionLoading: 'Loading latest submission...',
    submissionLoadFailed: 'Failed to load submissions for this question.',
    resultIdle: 'Submit your code to see the latest judge result.',
    examplePlaceholder: 'Placeholder',
    submissionUnavailableMessage: 'Submission is not available in this environment yet.',
  };

  const INITIAL_DATA = {
    session: {
      lessonTitle: PLACEHOLDERS.lessonTitleLoading,
      studentName: '',
    },
    lesson: {
      id: null,
      title: '',
      description: '',
      questions: [],
    },
    initialSelectedQuestionId: null,
  };

  function nowTimeString() {
    return new Date().toLocaleTimeString('zh-CN', { hour12: false });
  }

  function formatStatusLabel(value, fallback = 'not started') {
    if (!value) return fallback;
    return String(value).replaceAll('_', ' ');
  }

  function formatQuestionMeta(question) {
    if (!question) return '—';
    if (question.detailLoading) return 'loading details...';
    if (question.detailError) return 'detail load failed';
    if (question.status) return formatStatusLabel(question.status);
    if (question.detailLoaded) return 'ready';
    return 'not loaded';
  }

  function getErrorMessage(error, fallback = 'Request failed.') {
    if (!error) return fallback;
    if (typeof error.message === 'string' && error.message.trim()) return error.message.trim();
    return fallback;
  }

  const apiRequest = createApiRequest({ onUnauthorized: redirectToLogin });

  function createStore(seed) {
    const session = clone(seed.session);
    const lesson = clone(seed.lesson);
    const selectedQuestionId = seed.initialSelectedQuestionId || lesson.questions[0]?.id || null;
    const selectedQuestion = findById(lesson.questions, selectedQuestionId);

    return {
      session,
      lesson,
      ui: {
        selectedQuestionId,
        editorValue: selectedQuestion ? selectedQuestion.starterCode || '' : '',
        editorUndoStack: [],
        modified: false,
        isSubmitting: false,
        isBootstrapping: true,
        bootError: '',
        pollTimer: null,
      },
    };
  }

  const store = createStore(INITIAL_DATA);

  function getQuestions() {
    return Array.isArray(store.lesson.questions) ? store.lesson.questions : [];
  }

  function getQuestionById(questionId) {
    return findById(getQuestions(), questionId);
  }

  function getSelectedQuestion() {
    return getQuestionById(store.ui.selectedQuestionId);
  }

  function getSelectedQuestionIndex() {
    return getQuestions().findIndex((question) => question.id === store.ui.selectedQuestionId);
  }

  function normalizeLessonQuestionSummary(item, index) {
    const lessonQuestionId = Number(item.lesson_question_id);
    const sortOrder = Number(item.sort_order) || index + 1;
    return {
      id: lessonQuestionId,
      lessonQuestionId,
      backendQuestionId: Number(item.question_id) || null,
      title: item.title || ('Question ' + sortOrder),
      sortOrder,
      badgeLabel: '#' + sortOrder,
      status: '',
      difficulty: '—',
      description: '',
      starterCode: '',
      detailLoaded: false,
      detailLoading: false,
      detailError: '',
      submissions: [],
      submissionsLoaded: false,
      submissionsLoading: false,
      submissionsError: '',
      lastSubmission: null,
    };
  }

  function hydrateLesson(lesson) {
    const questionSummaries = Array.isArray(lesson?.questions) ? lesson.questions : [];
    const normalizedQuestions = questionSummaries
      .slice()
      .sort((a, b) => (Number(a.sort_order) || 0) - (Number(b.sort_order) || 0) || (Number(a.lesson_question_id) || 0) - (Number(b.lesson_question_id) || 0))
      .map(normalizeLessonQuestionSummary);

    store.lesson = {
      id: lesson?.id || null,
      title: lesson?.title || PLACEHOLDERS.noLessonTitle,
      description: lesson?.description || '',
      questions: normalizedQuestions,
    };
    store.session.lessonTitle = store.lesson.title;
    store.ui.selectedQuestionId = normalizedQuestions[0]?.id || null;
    store.ui.editorValue = normalizedQuestions[0]?.starterCode || '';
    store.ui.editorUndoStack = [];
    store.ui.modified = false;
    store.ui.bootError = '';
  }

  function renderQuestionDescriptionSections(description) {
    return `<article class="markdown-body problem-markdown">${renderMarkdown(normalizeStringValue(description))}</article>`;
  }

  function applyQuestionDetail(question, detail) {
    question.backendQuestionId = Number(detail?.id) || question.backendQuestionId;
    question.lessonQuestionId = Number(detail?.lesson_question_id) || question.lessonQuestionId;
    question.title = detail?.title || question.title;
    question.description = normalizeStringValue(detail?.description);
    question.starterCode = detail?.starter_code || '';
    question.sortOrder = Number(detail?.sort_order) || question.sortOrder;
    question.badgeLabel = '#' + question.sortOrder;
    question.detailLoaded = true;
    question.detailLoading = false;
    question.detailError = '';
  }

  function buildSubmissionMessage(submission) {
    const queueStatus = submission?.queueStatus || '';
    const verdict = submission?.verdict || '';
    const errorMessage = submission?.errorMessage || '';

    if (queueStatus === 'pending') return 'Submission queued for judging.';
    if (queueStatus === 'judging') return 'Judging in progress.';
    if (verdict === 'accepted') return 'Accepted.';
    if (errorMessage) return errorMessage;
    if (verdict) return 'Judge result: ' + formatStatusLabel(verdict) + '.';
    if (queueStatus) return 'Submission status: ' + formatStatusLabel(queueStatus) + '.';
    return PLACEHOLDERS.resultIdle;
  }

  function normalizeSubmissionRecord(item) {
    const queueStatus = item?.status || '';
    const verdict = item?.verdict || '';
    const errorMessage = item?.error_message || '';
    const sourceCode = typeof item?.source_code === 'string'
      ? item.source_code
      : (typeof item?.sourceCode === 'string' ? item.sourceCode : '');

    return {
      id: Number(item?.id) || null,
      queueStatus,
      verdict,
      sourceCode,
      submittedAt: item?.submitted_at || '',
      finishedAt: item?.finished_at || '',
      stdoutBuffer: normalizeStdoutBuffer(item?.stdout_buffer),
      judgeReport: normalizeJudgeReport(item?.judge_report),
      errorMessage,
      message: buildSubmissionMessage({
        queueStatus,
        verdict,
        errorMessage,
      }),
    };
  }

  function applyQuestionSubmissions(question, items) {
    const submissions = Array.isArray(items) ? items.map(normalizeSubmissionRecord) : [];
    question.submissions = submissions;
    question.lastSubmission = submissions[0] || null;
    question.status = question.lastSubmission?.verdict || question.lastSubmission?.queueStatus || '';
    question.submissionsLoaded = true;
    question.submissionsLoading = false;
    question.submissionsError = '';
  }

  function getQuestionEditorDefault(question) {
    if (!question) return '';

    const latestSource = question.lastSubmission?.sourceCode || '';
    return latestSource || question.starterCode || '';
  }

  function upsertSubmission(question, submission) {
    if (!question || !submission) return;

    const current = Array.isArray(question.submissions) ? question.submissions.slice() : [];
    const matchIndex = current.findIndex((item) => item.id && submission.id && item.id === submission.id);

    if (matchIndex >= 0) {
      current[matchIndex] = {
        ...current[matchIndex],
        ...submission,
      };
    } else {
      current.unshift(submission);
    }

    question.submissions = current;
    question.lastSubmission = current[0] || null;
    question.status = question.lastSubmission?.verdict || question.lastSubmission?.queueStatus || '';
    question.submissionsLoaded = true;
    question.submissionsLoading = false;
    question.submissionsError = '';
  }

  function setNoLessonState(message) {
    store.lesson = {
      id: null,
      title: PLACEHOLDERS.noLessonTitle,
      description: '',
      questions: [],
    };
    store.session.lessonTitle = PLACEHOLDERS.noLessonTitle;
    store.ui.selectedQuestionId = null;
    store.ui.editorValue = '';
    store.ui.modified = false;
    store.ui.bootError = message;
  }

  function shouldPollSubmission(submission) {
    const queueStatus = submission?.queueStatus || '';
    return queueStatus === 'pending' || queueStatus === 'judging';
  }

  const dom = {
    root: document.body,
    ...collectElementsById(
      'lessonContextTitle',
      'problemList',
      'problemContent',
      'problemStateText',
      'editor',
      'editorGutter',
      'editorHighlight',
      'editorHighlightWrap',
      'modifiedLabel',
      'submitBtn',
      'resetBtn',
      'resultMeta',
      'judgeReportBlock',
      'statusLeft',
      'statusMiddle',
      'studentName',
      'statusRight',
      'workspace',
      'editorShell',
      'mainArea',
      'sidebarSplitter',
      'problemSplitter',
      'resultSplitter'
    ),
  };

  const sessionService = {
    async getCurrentUser() {
      const data = await apiRequest(APP_CONFIG.api.me);
      return data.user || null;
    },
  };

  const lessonService = {
    async getCurrentLesson() {
      const data = await apiRequest(APP_CONFIG.api.currentLesson);
      return data.lesson || null;
    },
    async getQuestionDetail(lessonQuestionId) {
      const data = await apiRequest(APP_CONFIG.api.questionDetail(lessonQuestionId));
      return data.question || null;
    },
  };

  const submissionService = {
    async getQuestionSubmissions(lessonQuestionId) {
      const data = await apiRequest(APP_CONFIG.api.questionSubmissions(lessonQuestionId));
      return Array.isArray(data.submissions) ? data.submissions : [];
    },
    async getSubmissionDetail(submissionId) {
      const data = await apiRequest(APP_CONFIG.api.submissionDetail(submissionId));
      return data.submission || null;
    },
    async createSubmission(lessonQuestionId, sourceCode) {
      try {
        const data = await apiRequest(APP_CONFIG.api.submissions, {
          method: 'POST',
          body: JSON.stringify({
            lesson_question_id: lessonQuestionId,
            source_code: sourceCode,
          }),
        });
        return {
          mode: 'real',
          submission: data.submission || {},
          message: 'Submission queued for judging.',
        };
      } catch (error) {
        if (error instanceof ApiError && (error.status === 501 || error.code === 'not_implemented')) {
          return {
            mode: 'placeholder',
            submission: {},
            message: PLACEHOLDERS.submissionUnavailableMessage,
          };
        }
        throw error;
      }
    },
  };

  function syncEditorScroll() {
    dom.editorGutter.scrollTop = dom.editor.scrollTop;
    dom.editorHighlightWrap.scrollTop = dom.editor.scrollTop;
    dom.editorHighlightWrap.scrollLeft = dom.editor.scrollLeft;
  }

  function renderLessonContext() {
    const title = store.session.lessonTitle || 'Current Lesson';
    dom.lessonContextTitle.textContent = title;
    dom.lessonContextTitle.title = title;
  }

  function renderStatusbar() {
    const selectedIndex = getSelectedQuestionIndex();
    const questionCount = getQuestions().length;
    const selectedQuestion = getSelectedQuestion();
    const submission = selectedQuestion?.lastSubmission || null;
    const verdict = submission?.verdict || '';
    const queueStatus = submission?.queueStatus || 'idle';

    dom.statusMiddle.textContent = 'Problem ' + (selectedIndex >= 0 ? selectedIndex + 1 : 0) + ' / ' + questionCount;
    dom.studentName.textContent = 'Student: ' + (store.session.studentName || '—');
    dom.studentName.title = store.session.studentName || '';
    dom.statusRight.textContent = APP_CONFIG.languageLabel;

    if (store.ui.isBootstrapping) {
      dom.statusLeft.textContent = 'Loading...';
      return;
    }

    if (store.ui.bootError && questionCount === 0) {
      dom.statusLeft.textContent = 'No lesson';
      return;
    }

    if (store.ui.isSubmitting) {
      dom.statusLeft.textContent = 'Submitting...';
      return;
    }

    if (queueStatus === 'pending') {
      dom.statusLeft.textContent = 'Pending';
      return;
    }

    if (queueStatus === 'judging') {
      dom.statusLeft.textContent = 'Judging...';
      return;
    }

    if (verdict === 'accepted') {
      dom.statusLeft.textContent = 'Accepted';
      return;
    }

    if (selectedQuestion?.detailLoading) {
      dom.statusLeft.textContent = 'Loading detail...';
      return;
    }

    dom.statusLeft.textContent = store.ui.modified ? 'Modified' : 'Ready';
  }

  function renderQuestionList() {
    const questions = getQuestions();
    if (!questions.length) {
      dom.problemList.innerHTML = ui.emptyState(store.ui.bootError || PLACEHOLDERS.noQuestions);
      return;
    }

    dom.problemList.innerHTML = questions.map((question) => ui.listItem({
      className: 'problem-item',
      active: question.id === store.ui.selectedQuestionId,
      attrs: {
        'data-id': question.id,
        'data-state': question.status || ''
      },
      bodyHtml: `
          <span class="problem-item__dot"></span>
          <span class="problem-item__body">
            <div class="problem-item__title">${escapeHtml(question.title)}</div>
            <div class="problem-item__meta">${escapeHtml(formatQuestionMeta(question))}</div>
          </span>
          <span class="pill problem-item__badge">${escapeHtml(question.badgeLabel || '—')}</span>
      `
    })).join('');
  }

  function renderProblemPane() {
    const question = getSelectedQuestion();

    if (!question) {
      dom.problemStateText.textContent = 'idle';
      dom.problemContent.innerHTML = ui.emptyState(store.ui.bootError || PLACEHOLDERS.noQuestions, 'empty-state--boxed');
      return;
    }

    dom.problemStateText.textContent = formatQuestionMeta(question);

    if (question.detailLoading) {
      dom.problemContent.innerHTML = ui.problemSection(
        'Description',
        `<p class="problem-description">${escapeHtml(PLACEHOLDERS.questionDetailLoading)}</p>`
      );
      return;
    }

    if (question.detailError) {
      dom.problemContent.innerHTML = ui.problemSection(
        'Description',
        `<p class="problem-description">${escapeHtml(question.detailError)}</p>`
      );
      return;
    }

    if (!question.detailLoaded) {
      dom.problemContent.innerHTML = ui.problemSection(
        'Description',
        `<p class="problem-description">${escapeHtml(PLACEHOLDERS.questionDetailUnavailable)}</p>`
      );
      return;
    }

    dom.problemContent.innerHTML = `
      ${renderQuestionDescriptionSections(question.description)}
    `;
  }

  function renderEditorPane() {
    const editorValue = String(store.ui.editorValue ?? '');
    const highlightHtml = highlightLua(editorValue) + (editorValue.endsWith('\n') ? '\n ' : '');

    if (dom.editor.value !== editorValue) {
      dom.editor.value = editorValue;
    }

    dom.modifiedLabel.textContent = store.ui.modified ? 'Modified' : 'Saved';
    dom.editorGutter.textContent = buildLineNumbers(editorValue);
    dom.editorHighlight.innerHTML = highlightHtml;
    syncEditorScroll();
  }

  function renderResultPane() {
    const question = getSelectedQuestion();

    if (!question) {
      dom.resultMeta.textContent = 'No submission yet';
      dom.judgeReportBlock.innerHTML = renderJudgeReportPlaceholder('-');
      return;
    }

    if (question.submissionsLoading && !question.lastSubmission) {
      dom.resultMeta.textContent = 'Loading submissions...';
      dom.judgeReportBlock.innerHTML = renderJudgeReportPlaceholder('Loading judge report...');
      return;
    }

    if (question.submissionsError && !question.lastSubmission) {
      dom.resultMeta.textContent = 'Submission load failed';
      dom.judgeReportBlock.innerHTML = renderJudgeReportPlaceholder('Failed to load judge report.');
      return;
    }

    const submission = question.lastSubmission || null;

    if (!submission) {
      dom.resultMeta.textContent = 'No submission yet';
      dom.judgeReportBlock.innerHTML = renderJudgeReportPlaceholder(PLACEHOLDERS.resultIdle);
      return;
    }

    dom.resultMeta.textContent = buildJudgeResultMeta(submission);
    dom.judgeReportBlock.innerHTML = renderJudgeReportTerminal(submission.judgeReport || null, submission);
  }

  function renderActionState() {
    const selectedQuestion = getSelectedQuestion();
    const editorLocked = store.ui.isBootstrapping || !selectedQuestion || selectedQuestion.detailLoading || !selectedQuestion.detailLoaded;
    dom.editor.disabled = editorLocked;
    dom.submitBtn.disabled = store.ui.isSubmitting || editorLocked;
    dom.resetBtn.disabled = store.ui.isSubmitting || editorLocked;
  }

  function renderApp() {
    renderLessonContext();
    renderStatusbar();
    renderQuestionList();
    renderProblemPane();
    renderEditorPane();
    renderResultPane();
    renderActionState();
  }

  function stopPolling() {
    if (store.ui.pollTimer) {
      clearTimeout(store.ui.pollTimer);
      store.ui.pollTimer = null;
    }
  }

  function scheduleSubmissionPoll(questionId, delayMs = 2000) {
    stopPolling();

    const question = getQuestionById(questionId);
    if (!question || store.ui.selectedQuestionId !== questionId) return;

    const submission = question.lastSubmission || null;
    if (!shouldPollSubmission(submission)) return;

    store.ui.pollTimer = window.setTimeout(() => {
      store.ui.pollTimer = null;

      if (submission?.id) {
        void loadSubmissionDetail(questionId, submission.id, { render: true, schedule: true });
        return;
      }

      void loadQuestionSubmissions(questionId, { force: true, render: true });
    }, delayMs);
  }

  async function loadSubmissionDetail(lessonQuestionId, submissionId, options = {}) {
    const question = getQuestionById(lessonQuestionId);
    if (!question || !submissionId) return;

    try {
      const detail = await submissionService.getSubmissionDetail(submissionId);
      if (detail) {
        upsertSubmission(question, normalizeSubmissionRecord(detail));
        if (store.ui.selectedQuestionId === lessonQuestionId && !store.ui.modified) {
          store.ui.editorValue = getQuestionEditorDefault(question);
        }
      }
    } catch (error) {
      if (!question.lastSubmission) {
        question.submissionsError = getErrorMessage(error, 'Failed to load submission detail.');
      }
    }

    if (options.render !== false) {
      renderApp();
    }

    if (store.ui.selectedQuestionId === lessonQuestionId && options.schedule !== false) {
      scheduleSubmissionPoll(lessonQuestionId);
    }
  }

  async function loadQuestionSubmissions(lessonQuestionId, options = {}) {
    const question = getQuestionById(lessonQuestionId);
    if (!question || question.submissionsLoading) return;
    if (question.submissionsLoaded && options.force !== true) {
      if (store.ui.selectedQuestionId === lessonQuestionId) {
        scheduleSubmissionPoll(lessonQuestionId);
      }
      return;
    }

    question.submissionsLoading = true;
    question.submissionsError = '';
    if (options.render !== false) {
      renderApp();
    }

    try {
      const submissions = await submissionService.getQuestionSubmissions(lessonQuestionId);
      applyQuestionSubmissions(question, submissions);

      if (question.lastSubmission?.id) {
        await loadSubmissionDetail(lessonQuestionId, question.lastSubmission.id, {
          render: false,
          schedule: false,
        });
      }
    } catch (error) {
      question.submissionsLoading = false;
      question.submissionsError = getErrorMessage(error, PLACEHOLDERS.submissionLoadFailed);
    }

    if (options.render !== false) {
      renderApp();
    }

    if (store.ui.selectedQuestionId === lessonQuestionId) {
      scheduleSubmissionPoll(lessonQuestionId);
    }
  }

  async function loadQuestionDetail(lessonQuestionId, options = {}) {
    const question = getQuestionById(lessonQuestionId);
    if (!question || question.detailLoaded || question.detailLoading) return;

    question.detailLoading = true;
    question.detailError = '';
    renderApp();

    try {
      const detail = await lessonService.getQuestionDetail(lessonQuestionId);
      applyQuestionDetail(question, detail || {});

      if (store.ui.selectedQuestionId === lessonQuestionId && !store.ui.modified) {
        store.ui.editorValue = getQuestionEditorDefault(question);
      }
    } catch (error) {
      question.detailLoading = false;
      question.detailError = getErrorMessage(error, 'Failed to load question detail.');
    }

    if (options.render !== false) {
      renderApp();
    }
  }

  async function selectQuestion(questionId) {
    if (store.ui.isSubmitting) return;

    const question = getQuestionById(questionId);
    if (!question) return;

    stopPolling();
    store.ui.selectedQuestionId = questionId;
    store.ui.editorValue = question.detailLoaded ? getQuestionEditorDefault(question) : '';
    store.ui.editorUndoStack = [];
    store.ui.modified = false;
    renderApp();

    await Promise.all([
      question.detailLoaded ? Promise.resolve() : loadQuestionDetail(questionId, { render: false }),
      loadQuestionSubmissions(questionId, { render: false }),
    ]);
    renderApp();
  }

  function updateEditorValue(nextValue) {
    store.ui.editorValue = nextValue;
    store.ui.modified = true;
    renderEditorPane();
    renderResultPane();
    renderStatusbar();
  }

  function pushEditorUndoSnapshot(textarea) {
    const snapshot = {
      value: textarea.value,
      selectionStart: textarea.selectionStart,
      selectionEnd: textarea.selectionEnd,
      modified: store.ui.modified,
    };
    const lastSnapshot = store.ui.editorUndoStack[store.ui.editorUndoStack.length - 1];
    if (
      lastSnapshot
      && lastSnapshot.value === snapshot.value
      && lastSnapshot.selectionStart === snapshot.selectionStart
      && lastSnapshot.selectionEnd === snapshot.selectionEnd
      && lastSnapshot.modified === snapshot.modified
    ) {
      return;
    }

    store.ui.editorUndoStack.push(snapshot);
    if (store.ui.editorUndoStack.length > 100) {
      store.ui.editorUndoStack.shift();
    }
  }

  function undoEditorChange(textarea) {
    const snapshot = store.ui.editorUndoStack.pop();
    if (!snapshot) return false;

    store.ui.editorValue = snapshot.value;
    store.ui.modified = snapshot.modified;
    renderEditorPane();
    renderResultPane();
    renderStatusbar();
    textarea.focus();
    textarea.setSelectionRange(snapshot.selectionStart, snapshot.selectionEnd);
    return true;
  }

  function resetEditor() {
    const question = getSelectedQuestion();
    if (!question) return;

    pushEditorUndoSnapshot(dom.editor);
    store.ui.editorValue = getQuestionEditorDefault(question);
    store.ui.modified = false;
    renderEditorPane();
    renderResultPane();
    renderStatusbar();
  }

  const EDITOR_INDENT = '    ';

  function replaceEditorSelection(textarea, insertion, start = textarea.selectionStart, end = textarea.selectionEnd) {
    pushEditorUndoSnapshot(textarea);
    textarea.value = textarea.value.slice(0, start) + insertion + textarea.value.slice(end);
    textarea.selectionStart = start + insertion.length;
    textarea.selectionEnd = start + insertion.length;
    updateEditorValue(textarea.value);
  }

  function getCurrentLineBounds(value, cursor) {
    const lineStart = value.lastIndexOf('\n', Math.max(0, cursor - 1)) + 1;
    const nextNewlineIndex = value.indexOf('\n', cursor);
    const lineEnd = nextNewlineIndex === -1 ? value.length : nextNewlineIndex;
    return { lineStart, lineEnd };
  }

  function getAutoIndentation(value, cursor) {
    const { lineStart, lineEnd } = getCurrentLineBounds(value, cursor);
    const currentLine = value.slice(lineStart, lineEnd);
    return (currentLine.match(/^[\t ]*/) || [''])[0];
  }

  function handleEditorKeydown(event) {
    const textarea = event.target;
    const isUndoShortcut = event.key.toLowerCase() === 'z'
      && !event.shiftKey
      && !event.altKey
      && (event.ctrlKey || event.metaKey);

    if (isUndoShortcut) {
      event.preventDefault();
      undoEditorChange(textarea);
      return;
    }

    if (event.key === 'Tab') {
      event.preventDefault();
      replaceEditorSelection(textarea, EDITOR_INDENT);
      return;
    }

    if (event.key !== 'Enter') return;

    event.preventDefault();
    replaceEditorSelection(textarea, '\n' + getAutoIndentation(textarea.value, textarea.selectionStart));
  }

  async function submitCurrentCode() {
    const question = getSelectedQuestion();
    if (!question) return;

    store.ui.isSubmitting = true;
    renderActionState();
    renderStatusbar();

    try {
      const result = await submissionService.createSubmission(question.lessonQuestionId, store.ui.editorValue);
      const createdSubmission = result.submission || {};
      const normalizedSubmission = normalizeSubmissionRecord({
        ...createdSubmission,
        stdout_buffer: createdSubmission.stdout_buffer ?? null,
        judge_report: createdSubmission.judge_report || null,
      });

      normalizedSubmission.message = result.message;
      upsertSubmission(question, normalizedSubmission);
      store.ui.modified = false;
      renderApp();

      if (result.mode === 'real') {
        if (normalizedSubmission.id) {
          await loadSubmissionDetail(question.lessonQuestionId, normalizedSubmission.id, {
            render: false,
            schedule: false,
          });
          renderApp();
        }
        scheduleSubmissionPoll(question.lessonQuestionId, 1200);
      }
    } catch (error) {
      question.lastSubmission = {
        id: null,
        queueStatus: 'finished',
        verdict: 'system_error',
        submittedAt: nowTimeString(),
        finishedAt: nowTimeString(),
        message: 'Submission request failed.',
        stdoutBuffer: null,
        judgeReport: getErrorMessage(error, 'Unknown submission error.'),
        errorMessage: '',
      };
      question.status = question.lastSubmission.verdict || question.lastSubmission.queueStatus;
      renderApp();
    } finally {
      store.ui.isSubmitting = false;
      renderActionState();
      renderStatusbar();
    }
  }

  function bindEvents() {
    dom.problemList.addEventListener('click', (event) => {
      const button = event.target.closest('.problem-item');
      if (!button) return;
      void selectQuestion(Number(button.dataset.id));
    });

    dom.editor.addEventListener('input', (event) => {
      updateEditorValue(event.target.value);
    });

    dom.editor.addEventListener('beforeinput', (event) => {
      if (event.inputType === 'historyUndo') return;
      pushEditorUndoSnapshot(event.target);
    });

    dom.editor.addEventListener('keydown', handleEditorKeydown);
    dom.editor.addEventListener('scroll', syncEditorScroll);
    dom.submitBtn.addEventListener('click', submitCurrentCode);
    dom.resetBtn.addEventListener('click', resetEditor);

    bindResizablePane(dom.sidebarSplitter, {
      axis: 'x',
      ...APP_CONFIG.resizer.sidebar,
      container: dom.workspace,
      root: dom.root,
      breakpoint: APP_CONFIG.mobileBreakpoint,
    });

    bindResizablePane(dom.problemSplitter, {
      axis: 'x',
      ...APP_CONFIG.resizer.problem,
      container: dom.editorShell,
      root: dom.root,
      breakpoint: APP_CONFIG.mobileBreakpoint,
    });

    bindResizablePane(dom.resultSplitter, {
      axis: 'y',
      ...APP_CONFIG.resizer.result,
      container: dom.mainArea,
      root: dom.root,
      direction: -1,
      breakpoint: APP_CONFIG.mobileBreakpoint,
    });
  }

  async function bootstrap() {
    store.ui.isBootstrapping = true;
    store.ui.bootError = '';
    renderApp();

    try {
      const user = await sessionService.getCurrentUser();
      store.session.studentName = user?.username || '';
    } catch (error) {
      if (error instanceof ApiError && (error.status === 401 || error.status === 403)) {
        redirectToLogin();
        return;
      }
      setNoLessonState(getErrorMessage(error, 'Failed to load current user.'));
      store.ui.isBootstrapping = false;
      renderApp();
      return;
    }

    try {
      const lesson = await lessonService.getCurrentLesson();
      hydrateLesson(lesson || {});
    } catch (error) {
      if (error instanceof ApiError && (error.status === 401 || error.status === 403)) {
        redirectToLogin();
        return;
      }

      if (error instanceof ApiError && error.status === 404) {
        setNoLessonState('No current lesson has been assigned yet.');
        store.ui.isBootstrapping = false;
        renderApp();
        return;
      }

      setNoLessonState(getErrorMessage(error, 'Failed to load current lesson.'));
      store.ui.isBootstrapping = false;
      renderApp();
      return;
    }

    store.ui.isBootstrapping = false;
    renderApp();

    const selectedQuestion = getSelectedQuestion();
    if (selectedQuestion) {
      await Promise.all([
        loadQuestionDetail(selectedQuestion.id, { render: false }),
        loadQuestionSubmissions(selectedQuestion.id, { render: false }),
      ]);
      renderApp();
    }
  }

  function init() {
    bindEvents();
    renderApp();
    void bootstrap();
  }

  init();
})();
