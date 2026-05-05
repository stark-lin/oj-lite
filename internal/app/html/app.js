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

  function clamp(value, min, max) {
    return Math.min(Math.max(value, min), max);
  }

  function deepClone(value) {
    return JSON.parse(JSON.stringify(value));
  }

  function findById(items, id) {
    return items.find((item) => item.id === id) || null;
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

  function formatQuestionDescriptionValue(value) {
    if (typeof value === 'string') return value;
    if (value == null) return '—';
    return JSON.stringify(value, null, 2);
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

  function classNames(...values) {
    return values
      .reduce((items, value) => {
        if (!value) return items;
        return items.concat(String(value).split(/\s+/));
      }, [])
      .filter(Boolean)
      .join(' ');
  }

  function renderAttrs(attrs = {}) {
    return Object.entries(attrs)
      .filter(([, value]) => value !== false && value !== null && value !== undefined)
      .map(([key, value]) => value === true ? ` ${key}` : ` ${key}="${escapeHtml(value)}"`)
      .join('');
  }

  function mergeClassAttrs(attrs = {}, ...classes) {
    const nextAttrs = { ...attrs };
    const className = classNames(...classes, attrs.class);
    if (className) nextAttrs.class = className;
    return nextAttrs;
  }

  function emptyState(message, extraClass = '') {
    return `<div class="${classNames('empty-state', extraClass)}">${escapeHtml(message)}</div>`;
  }

  function paneMeta(value, attrs = {}, extraClass = '') {
    return `<span${renderAttrs(mergeClassAttrs(attrs, 'pane-meta', extraClass))}>${escapeHtml(value)}</span>`;
  }

  function pill(value, className = 'pill') {
    return `<span class="${classNames(className)}">${escapeHtml(value)}</span>`;
  }

  function statusPillClass(status) {
    if (status === 'active') return 'pill pill--success';
    if (status === 'disabled') return 'pill pill--warning';
    return 'pill';
  }

  function verdictLabel(verdict) {
    if (!verdict) return '—';
    return verdict.replaceAll('_', ' ');
  }

  function verdictPillClass(verdict) {
    if (verdict === 'accepted') return 'pill pill--success';
    if (verdict === 'wrong_answer' || verdict === 'pending') return 'pill pill--warning';
    if (verdict === 'runtime_error') return 'pill pill--danger';
    return 'pill';
  }

  function latestVerdictLabel(latest) {
    if (!latest) return 'No submission';
    return verdictLabel(latest.verdict || latest.status || '');
  }

  function latestVerdictPillClass(latest) {
    if (!latest) return 'pill';
    return verdictPillClass(latest.verdict || latest.status || '');
  }

  function actionButton(label, attrs = {}, extraClass = '') {
    return `<button class="${classNames('action-button', extraClass)}"${renderAttrs({ type: 'button', ...attrs })}>${escapeHtml(label)}</button>`;
  }

  function input(attrs = {}, extraClass = '') {
    return `<input${renderAttrs(mergeClassAttrs(attrs, extraClass))} />`;
  }

  function passwordInput(attrs = {}, extraClass = '') {
    return input({ type: 'password', autocomplete: 'new-password', ...attrs }, extraClass);
  }

  function loginInput(attrs = {}) {
    return input({ type: 'text', autocomplete: 'username', ...attrs }, 'input');
  }

  function loginPasswordInput(attrs = {}) {
    return passwordInput({ autocomplete: 'current-password', ...attrs }, 'input');
  }

  function loginField(options = {}) {
    const inputAttrs = options.inputAttrs || {};
    const inputHtml = options.inputHtml !== undefined
      ? options.inputHtml
      : (options.password ? loginPasswordInput(inputAttrs) : loginInput(inputAttrs));

    return `
      <label${renderAttrs(mergeClassAttrs(options.attrs || {}, 'field', options.extraClass || ''))}>
        <span class="label">${escapeHtml(options.label || '')}</span>
        ${inputHtml}
      </label>
    `;
  }

  function textarea(value = '', attrs = {}, extraClass = '') {
    return `<textarea${renderAttrs(mergeClassAttrs(attrs, extraClass))}>${escapeHtml(value)}</textarea>`;
  }

  function headerActions(content) {
    const body = Array.isArray(content) ? content.filter(Boolean).join('') : content;
    return `<div class="header-actions">${body || ''}</div>`;
  }

  function inlineActions(content) {
    const body = Array.isArray(content) ? content.filter(Boolean).join('') : content;
    return `<div class="inline-actions">${body || ''}</div>`;
  }

  function listTitle(value, attrs = {}, extraClass = '') {
    return `<div${renderAttrs(mergeClassAttrs(attrs, 'list-title', extraClass))}>${escapeHtml(value)}</div>`;
  }

  function listMeta(value, attrs = {}, extraClass = '') {
    return `<div${renderAttrs(mergeClassAttrs(attrs, 'list-meta', extraClass))}>${escapeHtml(value)}</div>`;
  }

  function listRow(leftHtml, rightHtml = '') {
    return `
      <div class="list-row">
        ${leftHtml}
        ${rightHtml}
      </div>
    `;
  }

  function listItem(options = {}) {
    const tagName = options.tag || 'button';
    const attrs = tagName === 'button'
      ? { type: 'button', ...(options.attrs || {}) }
      : (options.attrs || {});
    const titleHtml = options.titleHtml ?? listTitle(options.title || '');
    const rowHtml = options.rightHtml ? listRow(titleHtml, options.rightHtml) : titleHtml;
    const metaHtml = options.metaHtml !== undefined
      ? `<div class="list-meta">${options.metaHtml}</div>`
      : (options.meta !== undefined && options.meta !== null && options.meta !== '' ? listMeta(options.meta) : '');
    const bodyHtml = options.bodyHtml || `${rowHtml}${metaHtml}`;

    return `
      <${tagName} class="${classNames(options.className || 'list-button', options.active ? 'is-active' : '')}"${renderAttrs(attrs)}>
        ${bodyHtml}
      </${tagName}>
    `;
  }

  function detailBox(label, value, extraClass = '') {
    return `
      <div class="detail-box">
        <div class="detail-box__label">${escapeHtml(label)}</div>
        <div class="${classNames('detail-box__value', extraClass)}">${value}</div>
      </div>
    `;
  }

  function detailGrid(entries) {
    return `
      <div class="detail-grid">
        ${entries.map(([label, value, extraClass = '']) => detailBox(label, value, extraClass)).join('')}
      </div>
    `;
  }

  function stack(content = '', attrs = {}, extraClass = '') {
    const body = Array.isArray(content) ? content.filter(Boolean).join('') : content;
    return `<div${renderAttrs(mergeClassAttrs(attrs, 'stack', extraClass))}>${body || ''}</div>`;
  }

  function detailBlock(title, bodyHtml, options = {}) {
    return `
      <div class="detail-block">
        <div class="detail-block__title">${escapeHtml(title)}</div>
        <div class="${classNames('detail-block__content', options.prose ? 'detail-block__content--prose' : '', options.extraClass || '')}">${bodyHtml || ''}</div>
      </div>
    `;
  }

  function luaCodeBlock(source, options = {}) {
    const code = String(source || '');
    const highlighted = highlightLua(code) + (options.padTrailingNewline && code.endsWith('\n') ? '\n ' : '');

    return `
      <div class="${classNames('code-surface', options.extraClass || '')}">
        <div class="editor-gutter">${escapeHtml(buildLineNumbers(code))}</div>
        <pre class="code-highlight"><code>${highlighted}</code></pre>
      </div>
    `;
  }

  function section(options = {}) {
    const metaHtml = options.metaHtml !== undefined
      ? options.metaHtml
      : (options.meta !== undefined ? paneMeta(options.meta) : '');

    return `
      <section${renderAttrs(mergeClassAttrs(options.attrs || {}, 'section', options.extraClass || ''))}>
        <div class="section__header">
          <h3 class="section__title">${escapeHtml(options.title || '')}</h3>
          ${metaHtml}
        </div>
        <div${renderAttrs(mergeClassAttrs(options.bodyAttrs || {}, 'section__body', options.bodyClass || ''))}>
          ${options.body || ''}
        </div>
      </section>
    `;
  }

  function contentSizeClass(count) {
    if (count <= 0) return 'content-section--empty';
    if (count <= 2) return 'content-section--short';
    if (count <= 6) return 'content-section--medium';
    return 'content-section--long';
  }

  function applyContentSectionState(element, count) {
    if (!element) return;
    const contentCount = Number.isFinite(Number(count)) ? Number(count) : 0;
    element.dataset.contentCount = String(contentCount);
    element.classList.add('content-section');
    element.classList.remove(
      'content-section--empty',
      'content-section--short',
      'content-section--medium',
      'content-section--long'
    );
    element.classList.add(contentSizeClass(contentCount));
  }

  function contentSection(options = {}) {
    const items = Array.isArray(options.items) ? options.items.filter(Boolean) : null;
    const contentCount = Number.isFinite(Number(options.contentCount))
      ? Number(options.contentCount)
      : (items ? items.length : (options.body ? String(options.body).trim().length : 0));
    const body = items
      ? (items.length ? stack(items, options.stackAttrs || {}, options.stackClass || '') : emptyState(options.emptyText || 'No items yet.'))
      : (options.body || (options.emptyText ? emptyState(options.emptyText) : ''));

    return section({
      ...options,
      attrs: {
        ...(options.attrs || {}),
        'data-content-count': contentCount
      },
      extraClass: classNames('content-section', contentSizeClass(contentCount), options.extraClass || ''),
      body
    });
  }

  function formRow(options = {}) {
    const actionHtml = options.actionHtml !== undefined ? options.actionHtml : '<div></div>';
    return `
      <div class="form-row">
        <div class="detail-box__label">${escapeHtml(options.label || '')}</div>
        ${options.inputHtml || ''}
        ${actionHtml}
      </div>
    `;
  }

  function fieldActionRow(options = {}) {
    const actionHtml = options.actionHtml !== undefined ? options.actionHtml : '<div></div>';
    return `
      <div class="form-row form-row--field">
        ${loginField(options)}
        ${actionHtml}
      </div>
    `;
  }

  function problemSection(title, bodyHtml) {
    return `
      <section class="problem-section">
        <h3 class="problem-section__title">${escapeHtml(title)}</h3>
        ${bodyHtml || ''}
      </section>
    `;
  }

  function createColumnResizer(options = {}) {
    const workspace = options.workspace;
    const dividers = options.dividers || [];
    const widths = options.widths || {};
    const defaultWidths = options.defaultWidths || {};
    const minWidths = options.minWidths || {};
    const columnCount = Number(options.columnCount) || Object.keys(minWidths).length;
    const fixedCount = Math.max(0, columnCount - 1);
    const breakpoint = Number(options.breakpoint) || 1100;

    function dividerSize() {
      return parseFloat(getComputedStyle(document.documentElement).getPropertyValue('--divider-size')) || 1;
    }

    function getAvailableContentWidth() {
      return workspace.clientWidth - dividerSize() * Math.max(0, columnCount - 1);
    }

    function laterMinimumWidth(startIndex) {
      let total = Number(minWidths[columnCount]) || 0;
      for (let index = startIndex + 1; index <= fixedCount; index += 1) {
        total += Number(minWidths[index]) || 0;
      }
      return total;
    }

    function normalizeWidths() {
      const total = getAvailableContentWidth();
      let used = 0;
      for (let index = 1; index <= fixedCount; index += 1) {
        const max = total - used - laterMinimumWidth(index);
        widths[index] = clamp(widths[index], minWidths[index], max);
        used += widths[index];
      }
    }

    function apply() {
      if (window.innerWidth <= breakpoint) {
        workspace.style.gridTemplateColumns = '';
        return;
      }

      normalizeWidths();
      const total = getAvailableContentWidth();
      const fixedWidths = Array.from({ length: fixedCount }, (_, index) => widths[index + 1]);
      const lastWidth = Math.max(Number(minWidths[columnCount]) || 0, total - fixedWidths.reduce((sum, value) => sum + value, 0));
      workspace.style.gridTemplateColumns = fixedWidths
        .concat(lastWidth)
        .map((value, index) => `${value}px${index < columnCount - 1 ? ' var(--divider-size)' : ''}`)
        .join(' ');
    }

    function reset() {
      Object.keys(widths).forEach((key) => { delete widths[key]; });
      Object.assign(widths, deepClone(defaultWidths));
      apply();
    }

    function bind() {
      dividers.forEach((divider) => {
        divider.addEventListener('pointerdown', (event) => {
          if (window.innerWidth <= breakpoint) return;

          const dividerIndex = Number(divider.dataset.divider);
          const total = getAvailableContentWidth();
          const startX = event.clientX;
          const startWidths = { ...widths };

          divider.setPointerCapture(event.pointerId);
          divider.classList.add('is-dragging');
          document.body.style.cursor = 'col-resize';
          document.body.style.userSelect = 'none';

          function onPointerMove(moveEvent) {
            const delta = moveEvent.clientX - startX;
            let reserved = Number(minWidths[columnCount]) || 0;
            for (let index = 1; index <= fixedCount; index += 1) {
              if (index !== dividerIndex) reserved += Number(startWidths[index]) || 0;
            }
            widths[dividerIndex] = clamp(startWidths[dividerIndex] + delta, minWidths[dividerIndex], total - reserved);
            apply();
          }

          function stopDragging() {
            divider.classList.remove('is-dragging');
            document.body.style.cursor = '';
            document.body.style.userSelect = '';
            divider.removeEventListener('pointermove', onPointerMove);
            divider.removeEventListener('pointerup', stopDragging);
            divider.removeEventListener('pointercancel', stopDragging);
            divider.removeEventListener('lostpointercapture', stopDragging);
          }

          divider.addEventListener('pointermove', onPointerMove);
          divider.addEventListener('pointerup', stopDragging);
          divider.addEventListener('pointercancel', stopDragging);
          divider.addEventListener('lostpointercapture', stopDragging);
        });

        divider.addEventListener('dblclick', reset);
      });

      window.addEventListener('resize', apply);
    }

    return { apply, bind, getAvailableContentWidth, normalizeWidths, reset };
  }

  const ui = {
    actionButton,
    applyContentSectionState,
    contentSection,
    detailBox,
    detailBlock,
    detailGrid,
    emptyState,
    formRow,
    headerActions,
    inlineActions,
    input,
    fieldActionRow,
    loginField,
    loginInput,
    loginPasswordInput,
    listItem,
    listMeta,
    listRow,
    listTitle,
    luaCodeBlock,
    paneMeta,
    passwordInput,
    pill,
    problemSection,
    section,
    stack,
    textarea
  };

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
    createColumnResizer,
    createApiRequest,
    deepClone,
    escapeHtml,
    findById,
    buildJudgeResultMeta,
    formatDateTime,
    formatJSONValue,
    formatQuestionDescriptionValue,
    formatStdoutBuffer,
    highlightLua,
    latestVerdictLabel,
    latestVerdictPillClass,
    normalizeJSONValue,
    normalizeJudgeReport,
    normalizeStdoutBuffer,
    parseJSONSafe,
    renderJudgeReportPlaceholder,
    renderJudgeReportTerminal,
    sortByIDAsc,
    sortBySortOrder,
    statusPillClass,
    verdictLabel,
    verdictPillClass,
    ui
  };
}());
