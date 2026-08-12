(function () {
  class ApiError extends Error {
    constructor(message, options = {}) {
      super(message);
      this.name = 'ApiError';
      this.status = options.status || 0;
      this.code = options.code || 'request_failed';
      this.details = options.details || null;
      this.requestId = options.requestId || null;
    }
  }

  function escapeHtml(value) {
    return String(value)
      .replaceAll('&', '&amp;')
      .replaceAll('<', '&lt;')
      .replaceAll('>', '&gt;')
      .replaceAll('"', '&quot;')
      .replaceAll("'", '&#39;');
  }

  function renderMarkdown(value) {
    const source = String(value ?? '');
    if (!source.trim()) {
      return '<p class="markdown-empty">—</p>';
    }

    const parse = window.marked?.parse;
    const sanitize = window.DOMPurify?.sanitize;
    if (typeof parse !== 'function' || typeof sanitize !== 'function') {
      return `<pre class="markdown-fallback">${escapeHtml(source)}</pre>`;
    }

    try {
      const rendered = parse(source, { async: false, gfm: true });
      return sanitize(rendered, { USE_PROFILES: { html: true } });
    } catch (error) {
      void error;
      return `<pre class="markdown-fallback">${escapeHtml(source)}</pre>`;
    }
  }

  function clamp(value, min, max) {
    return Math.min(Math.max(value, min), max);
  }

  function deepClone(value) {
    return JSON.parse(JSON.stringify(value));
  }

  function findById(items, id) {
    return items.find((item) => item.id === id) || null;
  }

  function collectElementsById(...ids) {
    return Object.fromEntries(ids.map((id) => [id, document.getElementById(id)]));
  }

  function sortByIDAsc(items) {
    return items.slice().sort((a, b) => (Number(a?.id) || 0) - (Number(b?.id) || 0));
  }

  function sortBySortOrder(items) {
    return items.slice().sort((a, b) => {
      const orderDiff = (Number(a?.sortOrder) || 0) - (Number(b?.sortOrder) || 0);
      if (orderDiff !== 0) return orderDiff;
      return (Number(a?.id) || 0) - (Number(b?.id) || 0);
    });
  }

  function normalizeJSONValue(value, fallback) {
    if (value == null) return fallback;
    if (typeof value === 'string') {
      const trimmed = value.trim();
      if (!trimmed) return fallback;
      try {
        return JSON.parse(trimmed);
      } catch (error) {
        void error;
        return fallback;
      }
    }
    if (typeof value === 'object') return value;
    return fallback;
  }

  function normalizeStringValue(value, fallback = '') {
    return typeof value === 'string' ? value : fallback;
  }

  async function parseJSONSafe(response) {
    const text = await response.text();
    if (!text) {
      return null;
    }

    try {
      return JSON.parse(text);
    } catch (error) {
      void error;
      return null;
    }
  }

  function applyAppName(value) {
    const appName = typeof value === 'string' ? value.trim() : '';
    const title = document.body?.dataset.appTitle || '';

    document.title = appName && title ? `${appName} · ${title}` : title;
    document.querySelectorAll('[data-app-name]').forEach((element) => {
      element.textContent = appName;
      element.hidden = !appName;
    });
  }

  async function loadAppName() {
    applyAppName('');

    try {
      const response = await fetch('/healthz', {
        credentials: 'same-origin',
        headers: { Accept: 'application/json' }
      });
      if (!response.ok) return;

      const payload = await parseJSONSafe(response);
      applyAppName(payload?.data?.service);
    } catch (error) {
      void error;
    }
  }

  function redirectToLogin() {
    window.location.href = '/';
  }

  function formatDateTime(value) {
    if (!value) return '—';
    const timestamp = Date.parse(value);
    if (Number.isNaN(timestamp)) return value;
    return new Date(timestamp).toLocaleString('zh-CN', {
      hour12: false,
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  function buildLineNumbers(value) {
    const lineCount = Math.max(1, String(value).split('\n').length);
    return Array.from({ length: lineCount }, (_, index) => String(index + 1)).join('\n');
  }

  const LUA_KEYWORDS = new Set([
    'and', 'break', 'do', 'else', 'elseif', 'end', 'for', 'function',
    'if', 'in', 'local', 'not', 'or', 'repeat', 'return', 'then',
    'until', 'while'
  ]);
  const LUA_CONSTANTS = new Set(['nil', 'true', 'false']);

  function codeToken(className, value) {
    return `<span class="${className}">${escapeHtml(value)}</span>`;
  }

  function isIdentifierStart(char) {
    if (!char) return false;
    const code = char.charCodeAt(0);
    return char === '_' || (code >= 65 && code <= 90) || (code >= 97 && code <= 122);
  }

  function isIdentifierPart(char) {
    if (!char) return false;
    const code = char.charCodeAt(0);
    return isIdentifierStart(char) || (code >= 48 && code <= 57);
  }

  function isDigit(char) {
    if (!char) return false;
    const code = char.charCodeAt(0);
    return code >= 48 && code <= 57;
  }

  function isWhitespace(char) {
    if (!char) return false;
    const code = char.charCodeAt(0);
    return code === 9 || code === 10 || code === 13 || code === 32;
  }

  function highlightLua(source) {
    const code = String(source || '');
    let html = '';
    let index = 0;
    let expectFunctionName = false;

    while (index < code.length) {
      const char = code[index];
      const nextChar = code[index + 1] || '';

      if (char === '-' && nextChar === '-') {
        let end = index + 2;
        while (end < code.length && code[end] !== '\n') end += 1;
        html += codeToken('tok-comment', code.slice(index, end));
        index = end;
        continue;
      }

      if (char === '"' || char === "'") {
        const quote = char;
        let end = index + 1;
        while (end < code.length) {
          if (code[end] === '\\') {
            end += 2;
            continue;
          }
          if (code[end] === quote) {
            end += 1;
            break;
          }
          end += 1;
        }
        html += codeToken('tok-string', code.slice(index, end));
        index = end;
        continue;
      }

      if (isDigit(char)) {
        let end = index + 1;
        while (end < code.length) {
          const current = code[end];
          if (isDigit(current) || current === '.' || current === '_') {
            end += 1;
            continue;
          }
          break;
        }
        html += codeToken('tok-number', code.slice(index, end));
        index = end;
        expectFunctionName = false;
        continue;
      }

      if (isIdentifierStart(char)) {
        let end = index + 1;
        while (end < code.length && isIdentifierPart(code[end])) end += 1;
        const word = code.slice(index, end);

        if (LUA_KEYWORDS.has(word)) {
          html += codeToken('tok-keyword', word);
          expectFunctionName = word === 'function';
        } else if (LUA_CONSTANTS.has(word)) {
          html += codeToken('tok-constant', word);
          expectFunctionName = false;
        } else if (expectFunctionName) {
          html += codeToken('tok-function', word);
          expectFunctionName = false;
        } else {
          html += escapeHtml(word);
        }

        index = end;
        continue;
      }

      if (!isWhitespace(char)) {
        expectFunctionName = false;
      }

      html += escapeHtml(char);
      index += 1;
    }

    return html || ' ';
  }

  function normalizeStdoutBuffer(value) {
    if (value == null || value === '') return null;

    const legacy = (text) => {
      const normalized = String(text || '');
      if (!normalized.trim()) return null;
      return {
        cases: [],
        legacyText: normalized
      };
    };

    if (typeof value === 'string') {
      const trimmed = value.trim();
      if (!trimmed) return null;

      try {
        return normalizeStdoutBuffer(JSON.parse(trimmed));
      } catch (error) {
        void error;
        return legacy(value);
      }
    }

    if (typeof value !== 'object') {
      return legacy(String(value));
    }

    if (typeof value?.legacyText === 'string' && value.legacyText.trim()) {
      return {
        cases: [],
        legacyText: value.legacyText
      };
    }

    const rawCases = Array.isArray(value?.cases) ? value.cases : [];
    const cases = rawCases
      .map((item, index) => {
        const stdout = typeof item?.stdout === 'string' ? item.stdout : '';
        if (!stdout) return null;

        const caseIndex = Number(item?.index);
        return {
          index: Number.isFinite(caseIndex) && caseIndex > 0 ? caseIndex : index + 1,
          stdout
        };
      })
      .filter(Boolean);

    if (!cases.length) return null;

    return {
      cases,
      legacyText: ''
    };
  }

  function formatStdoutBuffer(value) {
    const stdoutBuffer = normalizeStdoutBuffer(value);
    if (!stdoutBuffer) return '—';
    if (stdoutBuffer.legacyText) return stdoutBuffer.legacyText;

    return stdoutBuffer.cases
      .map((item) => `Case ${item.index}\n${item.stdout || '—'}`)
      .join('\n\n');
  }

  function formatJSONValue(value, fallback = '—') {
    if (value == null || value === '') return fallback;
    if (typeof value === 'string') {
      try {
        return JSON.stringify(JSON.parse(value), null, 2);
      } catch (error) {
        void error;
        return value;
      }
    }
    return JSON.stringify(value, null, 2);
  }

  function normalizeJudgeReport(value) {
    if (value == null || value === '') return null;

    if (typeof value === 'string') {
      const trimmed = value.trim();
      if (!trimmed) return null;

      try {
        return JSON.parse(trimmed);
      } catch (error) {
        void error;
        return value;
      }
    }

    return value;
  }

  const TERMINAL_LABEL_WIDTH = 11;
  const TERMINAL_EMPTY = '-';

  function termSpan(className, value) {
    return `<span class="${className}">${escapeHtml(value)}</span>`;
  }

  function normalizeTerminalStatus(value, fallback = 'FINISHED') {
    if (!value) return fallback;

    const normalized = String(value)
      .trim()
      .replace(/([a-z0-9])([A-Z])/g, '$1_$2')
      .replace(/[^A-Za-z0-9]+/g, '_')
      .replace(/^_+|_+$/g, '')
      .toUpperCase();

    return normalized || fallback;
  }

  function formatTerminalKey(value) {
    return normalizeTerminalStatus(value, 'FIELD');
  }

  function terminalLabelPrefix(label) {
    const normalized = formatTerminalKey(label).replace(/:+$/g, '') + ':';
    return normalized.length >= TERMINAL_LABEL_WIDTH
      ? normalized + ' '
      : normalized.padEnd(TERMINAL_LABEL_WIDTH, ' ');
  }

  function renderTermRule(label, fillChar = '=') {
    const prefix = String(label || '');
    const fill = String(fillChar || '=').slice(0, 1) || '=';
    return termSpan('term-rule', prefix + ' ' + fill.repeat(320));
  }

  function formatTerminalScalar(value) {
    if (value == null || value === '') return TERMINAL_EMPTY;
    if (typeof value === 'string') return value;
    if (typeof value === 'number' || typeof value === 'boolean') return String(value);

    try {
      const encoded = JSON.stringify(value);
      return encoded == null ? String(value) : encoded;
    } catch (error) {
      void error;
      return String(value);
    }
  }

  function formatTerminalValue(value, options = {}) {
    if (value == null || value === '') return TERMINAL_EMPTY;

    if (Array.isArray(value)) {
      if (!value.length) return TERMINAL_EMPTY;
      if (options.unwrapSingleArray && value.length === 1) {
        return formatTerminalValue(value[0], { ...options, unwrapSingleArray: false });
      }

      return value.map(formatTerminalScalar).join(', ');
    }

    if (typeof value === 'object') {
      try {
        const encoded = JSON.stringify(value);
        return encoded == null ? String(value) : encoded;
      } catch (error) {
        void error;
        return String(value);
      }
    }

    return formatTerminalScalar(value);
  }

  function renderTermField(label, value, className = 'term-value', options = {}) {
    const prefix = terminalLabelPrefix(label);
    const text = formatTerminalValue(value, options);
    const normalizedText = String(text || TERMINAL_EMPTY).replace(/[\r\n]+$/g, '');
    const lines = (normalizedText || TERMINAL_EMPTY).split('\n');

    return lines.map((line, index) => {
      const content = line || (index === 0 ? TERMINAL_EMPTY : '');
      if (index === 0) {
        return termSpan('term-key', prefix) + termSpan(className, content);
      }
      return ' '.repeat(prefix.length) + termSpan(className, content);
    }).join('\n');
  }

  function renderTerminalReport(lines) {
    return `<pre class="terminal-report">${lines.join('\n')}</pre>`;
  }

  function renderJudgeReportPlaceholder(message) {
    return renderTerminalReport([termSpan('term-muted', message || TERMINAL_EMPTY)]);
  }

  function isAcceptedJudgeStatus(value) {
    const normalized = String(value || '').trim().toLowerCase();
    return normalized === 'accepted'
      || normalized === 'pass'
      || normalized === 'passed'
      || normalized === 'ok'
      || normalized === 'success';
  }

  function isJudgeCasePassed(caseItem) {
    if (typeof caseItem?.comparison?.matched === 'boolean') return caseItem.comparison.matched;
    if (typeof caseItem?.passed === 'boolean') return caseItem.passed;
    if (typeof caseItem?.ok === 'boolean') return caseItem.ok;
    return isAcceptedJudgeStatus(caseItem?.verdict || caseItem?.status || caseItem?.result);
  }

  function getJudgeCases(report) {
    return report && typeof report === 'object' && Array.isArray(report.cases) ? report.cases : [];
  }

  function getJudgeCaseStats(report) {
    const cases = getJudgeCases(report);
    return {
      cases,
      passedCount: cases.filter(isJudgeCasePassed).length,
      totalCount: cases.length,
    };
  }

  function getTerminalStatusClass(value, fallbackPassed) {
    const normalized = String(value || '').trim().toLowerCase();

    if (isAcceptedJudgeStatus(normalized)) return 'term-pass';
    if (normalized === 'pending' || normalized === 'judging') return 'term-accent';
    if (
      normalized === 'failed'
      || normalized === 'fail'
      || normalized === 'wrong_answer'
      || normalized === 'runtime_error'
      || normalized === 'system_error'
      || normalized === 'error'
    ) {
      return 'term-fail';
    }

    if (typeof fallbackPassed === 'boolean') return fallbackPassed ? 'term-pass' : 'term-fail';
    return 'term-value';
  }

  function buildJudgeResultMeta(submission) {
    if (!submission) return 'No submission yet';

    const report = normalizeJudgeReport(submission?.judgeReport ?? submission?.judge_report ?? null);
    const reportObject = report && typeof report === 'object' ? report : null;
    const { passedCount, totalCount } = getJudgeCaseStats(reportObject);
    const status = normalizeTerminalStatus(
      submission?.verdict
      || reportObject?.verdict
      || reportObject?.status
      || submission?.queueStatus
      || submission?.status
      || 'finished'
    );

    if (totalCount > 0) {
      return status + ' · ' + passedCount + ' / ' + totalCount + ' PASSED';
    }

    return status;
  }

  function getJudgeCaseInput(caseItem) {
    return caseItem?.input ?? null;
  }

  function getJudgeCaseExpected(caseItem) {
    return caseItem?.reference?.returnValues ?? caseItem?.expected ?? caseItem?.expected_output ?? null;
  }

  function getJudgeCaseActual(caseItem) {
    return caseItem?.student?.returnValues ?? caseItem?.actual ?? caseItem?.output ?? caseItem?.actual_output ?? null;
  }

  function getJudgeCaseStdout(caseItem, submission, caseNumber, casePosition) {
    const directStdout = caseItem?.student?.stdoutBuffer ?? caseItem?.student?.stdout ?? caseItem?.stdout ?? '';
    if (typeof directStdout === 'string' && directStdout) return directStdout;

    const stdoutBuffer = submission?.stdoutBuffer || normalizeStdoutBuffer(submission?.stdout_buffer) || null;
    const stdoutCases = Array.isArray(stdoutBuffer?.cases) ? stdoutBuffer.cases : [];
    const matchedCase = stdoutCases.find((item) => Number(item?.index) === caseNumber);
    if (typeof matchedCase?.stdout === 'string') return matchedCase.stdout;

    const positionalCase = stdoutCases[casePosition];
    const positionalIndex = Number(positionalCase?.index);
    const canUsePosition = !Number.isFinite(positionalIndex) || positionalIndex <= 0;
    return canUsePosition && typeof positionalCase?.stdout === 'string' ? positionalCase.stdout : '';
  }

  function renderJudgeReportTerminal(reportValue, submission = {}) {
    const report = normalizeJudgeReport(reportValue);
    const reportObject = report && typeof report === 'object' ? report : null;
    const { cases, passedCount, totalCount } = getJudgeCaseStats(reportObject);
    const allPassed = totalCount > 0 && passedCount === totalCount;
    const statusRaw =
      submission?.verdict
      || reportObject?.verdict
      || reportObject?.status
      || submission?.queueStatus
      || submission?.status
      || '';
    const statusText = normalizeTerminalStatus(
      statusRaw || (totalCount > 0 ? (allPassed ? 'accepted' : 'wrong_answer') : 'finished')
    );

    const lines = [
      renderTermRule('== JUDGE RESULT', '='),
      renderTermField('VERDICT', statusText, getTerminalStatusClass(statusRaw || statusText, totalCount > 0 ? allPassed : undefined)),
    ];

    if (totalCount > 0) {
      lines.push(renderTermField('PASSED', passedCount + ' / ' + totalCount + ' cases', allPassed ? 'term-pass' : 'term-fail'));
    }

    if (submission?.submittedAt || submission?.submitted_at) {
      lines.push(renderTermField('SUBMITTED', formatDateTime(submission.submittedAt || submission.submitted_at)));
    }

    if (submission?.finishedAt || submission?.finished_at) {
      lines.push(renderTermField('FINISHED', formatDateTime(submission.finishedAt || submission.finished_at)));
    }

    if (submission?.errorMessage || submission?.error_message) {
      lines.push(renderTermField('MESSAGE', submission.errorMessage || submission.error_message, 'term-muted'));
    }

    if (!report) {
      lines.push('');
      lines.push(renderTermField('DETAILS', 'No judge details yet.', 'term-muted'));
      lines.push('');
      lines.push(renderTermRule('== END', '='));
      return renderTerminalReport(lines);
    }

    if (!reportObject) {
      lines.push('');
      lines.push(renderTermField('MESSAGE', report, 'term-muted'));
      lines.push('');
      lines.push(renderTermRule('== END', '='));
      return renderTerminalReport(lines);
    }

    if (!cases.length) {
      const extraFields = Object.entries(reportObject)
        .filter(([key]) => key !== 'cases')
        .filter(([, value]) => value != null && value !== '');

      lines.push('');

      if (extraFields.length) {
        extraFields.forEach(([key, value]) => {
          lines.push(renderTermField(key, value, 'term-value'));
        });
      } else {
        lines.push(renderTermField('DETAILS', 'No test case details.', 'term-muted'));
      }

      lines.push('');
      lines.push(renderTermRule('== END', '='));
      return renderTerminalReport(lines);
    }

    cases.forEach((caseItem, index) => {
      const passed = isJudgeCasePassed(caseItem);
      const caseNumber = Number(caseItem?.index) || index + 1;
      const caseStdout = getJudgeCaseStdout(caseItem, submission, caseNumber, index);
      const referenceError = caseItem?.reference?.errorMessage ?? caseItem?.reference?.error;
      const studentError = caseItem?.student?.errorMessage ?? caseItem?.student?.error;
      const reason = caseItem?.comparison?.reason ?? caseItem?.message ?? caseItem?.error_message ?? TERMINAL_EMPTY;

      lines.push('');
      lines.push(renderTermRule('-- CASE ' + String(caseNumber).padStart(2, '0'), '-'));
      lines.push(renderTermField('STATUS', passed ? 'PASSED' : 'FAILED', passed ? 'term-pass' : 'term-fail'));
      lines.push(renderTermField('INPUT', getJudgeCaseInput(caseItem)));
      lines.push(renderTermField('EXPECTED', getJudgeCaseExpected(caseItem), 'term-accent', { unwrapSingleArray: true }));
      lines.push(renderTermField('ACTUAL', getJudgeCaseActual(caseItem), passed ? 'term-pass' : 'term-fail', { unwrapSingleArray: true }));
      lines.push(renderTermField('REASON', reason, 'term-muted'));
      lines.push(renderTermField('STDOUT', caseStdout || TERMINAL_EMPTY, 'term-muted'));

      if (referenceError) {
        lines.push(renderTermField('REF_ERROR', referenceError, 'term-muted'));
      }

      if (studentError) {
        lines.push(renderTermField('RUN_ERROR', studentError, 'term-muted'));
      }
    });

    lines.push('');
    lines.push(renderTermRule('== END', '='));
    return renderTerminalReport(lines);
  }

  function createApiRequest(options = {}) {
    const credentials = options.credentials || 'same-origin';
    const onUnauthorized = options.onUnauthorized || null;

    return async function apiRequest(path, requestOptions = {}) {
      const response = await fetch(path, {
        method: requestOptions.method || 'GET',
        credentials,
        headers: {
          Accept: 'application/json',
          ...(requestOptions.body ? { 'Content-Type': 'application/json' } : {}),
          ...(requestOptions.headers || {})
        },
        body: requestOptions.body
      });

      const payload = await parseJSONSafe(response) || {};
      if (!response.ok) {
        if ((response.status === 401 || response.status === 403) && typeof onUnauthorized === 'function') {
          onUnauthorized();
        }

        const errorBody = payload?.error || {};
        throw new ApiError(errorBody.message || `Request failed with status ${response.status}.`, {
          status: response.status,
          code: errorBody.code || 'request_failed',
          details: errorBody.details || null,
          requestId: payload?.request_id || null
        });
      }

      return payload?.data || {};
    };
  }

  window.OJLite = {
    ApiError,
    buildLineNumbers,
    clamp,
    collectElementsById,
    createApiRequest,
    deepClone,
    escapeHtml,
    findById,
    buildJudgeResultMeta,
    formatDateTime,
    formatJSONValue,
    formatStdoutBuffer,
    highlightLua,
    normalizeJSONValue,
    normalizeJudgeReport,
    normalizeStringValue,
    normalizeStdoutBuffer,
    parseJSONSafe,
    renderMarkdown,
    redirectToLogin,
    renderJudgeReportPlaceholder,
    renderJudgeReportTerminal,
    sortByIDAsc,
    sortBySortOrder,
  };

  void loadAppName();
}());
