(() => {
  'use strict';

  const {
    buildLineNumbers,
    clamp,
    deepClone,
    escapeHtml,
    highlightLua
  } = window.OJLite;

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
    return `<div class="${classNames('surface empty-state', extraClass)}">${escapeHtml(message)}</div>`;
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

  function codeToolbar(content) {
    const body = Array.isArray(content) ? content.filter(Boolean).join('') : content;
    return `<div class="code-toolbar">${body || ''}</div>`;
  }

  function formActions(content) {
    const body = Array.isArray(content) ? content.filter(Boolean).join('') : content;
    return `<div class="form-actions">${body || ''}</div>`;
  }

  function formGrid(content) {
    const body = Array.isArray(content) ? content.filter(Boolean).join('') : content;
    return `<div class="form-grid">${body || ''}</div>`;
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
      <${tagName} class="${classNames('list-item', options.className || 'list-button', options.active ? 'is-active' : '')}"${renderAttrs(attrs)}>
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

    const headerHtml = options.header === false ? '' : `
      <div class="section__header">
        <h3 class="section__title">${escapeHtml(options.title || '')}</h3>
        ${metaHtml}
      </div>
    `;

    return `
      <section${renderAttrs(mergeClassAttrs(options.attrs || {}, 'surface section', options.extraClass || ''))}>
        ${headerHtml}
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
      <section class="surface problem-section">
        <h3 class="problem-section__title">${escapeHtml(title)}</h3>
        ${bodyHtml || ''}
      </section>
    `;
  }

  function createColumnResizer(options = {}) {
    const workspace = options.workspace;
    const resizers = options.resizers || [];
    const widths = options.widths || {};
    const defaultWidths = options.defaultWidths || {};
    const minWidths = options.minWidths || {};
    const columnCount = Number(options.columnCount) || Object.keys(minWidths).length;
    const fixedCount = Math.max(0, columnCount - 1);
    const breakpoint = Number(options.breakpoint) || 1100;

    function resizerSize() {
      return parseFloat(getComputedStyle(document.documentElement).getPropertyValue('--divider-size')) || 1;
    }

    function getAvailableContentWidth() {
      return workspace.clientWidth - resizerSize() * Math.max(0, columnCount - 1);
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
      resizers.forEach((resizer) => {
        resizer.addEventListener('pointerdown', (event) => {
          if (window.innerWidth <= breakpoint) return;

          event.preventDefault();
          const resizerIndex = Number(resizer.dataset.resizer);
          const total = getAvailableContentWidth();
          const startX = event.clientX;
          const startWidths = { ...widths };

          resizer.setPointerCapture(event.pointerId);
          resizer.classList.add('is-dragging');
          document.body.style.cursor = 'col-resize';
          document.body.style.userSelect = 'none';

          function onPointerMove(moveEvent) {
            const delta = moveEvent.clientX - startX;
            let reserved = Number(minWidths[columnCount]) || 0;
            for (let index = 1; index <= fixedCount; index += 1) {
              if (index !== resizerIndex) reserved += Number(startWidths[index]) || 0;
            }
            widths[resizerIndex] = clamp(startWidths[resizerIndex] + delta, minWidths[resizerIndex], total - reserved);
            apply();
          }

          function stopDragging() {
            resizer.classList.remove('is-dragging');
            document.body.style.cursor = '';
            document.body.style.userSelect = '';
            if (resizer.hasPointerCapture(event.pointerId)) {
              resizer.releasePointerCapture(event.pointerId);
            }
            window.removeEventListener('pointermove', onPointerMove);
            window.removeEventListener('pointerup', stopDragging);
            window.removeEventListener('pointercancel', stopDragging);
          }

          window.addEventListener('pointermove', onPointerMove);
          window.addEventListener('pointerup', stopDragging);
          window.addEventListener('pointercancel', stopDragging);
        });

        resizer.addEventListener('dblclick', reset);
      });

      window.addEventListener('resize', apply);
    }

    return { apply, bind, getAvailableContentWidth, normalizeWidths, reset };
  }

  function bindResizablePane(resizer, options = {}) {
    if (!resizer) return;

    const axis = options.axis === 'y' ? 'y' : 'x';
    const breakpoint = Number(options.breakpoint) || 0;
    const root = options.root || document.documentElement;

    resizer.addEventListener('pointerdown', (event) => {
      if (breakpoint && window.innerWidth <= breakpoint) return;

      event.preventDefault();
      resizer.classList.add('is-dragging');
      resizer.setPointerCapture(event.pointerId);
      document.body.style.cursor = axis === 'x' ? 'col-resize' : 'row-resize';
      document.body.style.userSelect = 'none';

      const containerRect = options.container.getBoundingClientRect();
      const startPosition = axis === 'x' ? event.clientX : event.clientY;
      const currentValue = parseFloat(getComputedStyle(root).getPropertyValue(options.cssVar)) || 0;
      const direction = Number(options.direction) || 1;
      const maximum = typeof options.max === 'function' ? options.max(containerRect) : Number(options.max);

      function onPointerMove(moveEvent) {
        const position = axis === 'x' ? moveEvent.clientX : moveEvent.clientY;
        const nextValue = clamp(currentValue + (position - startPosition) * direction, options.min, maximum);
        root.style.setProperty(options.cssVar, `${nextValue}px`);
      }

      function stopDragging() {
        resizer.classList.remove('is-dragging');
        document.body.style.cursor = '';
        document.body.style.userSelect = '';
        window.removeEventListener('pointermove', onPointerMove);
        window.removeEventListener('pointerup', stopDragging);
        window.removeEventListener('pointercancel', stopDragging);
      }

      window.addEventListener('pointermove', onPointerMove);
      window.addEventListener('pointerup', stopDragging, { once: true });
      window.addEventListener('pointercancel', stopDragging, { once: true });
    });
  }

  const ui = {
    actionButton,
    applyContentSectionState,
    contentSection,
    detailBox,
    detailBlock,
    detailGrid,
    emptyState,
    fieldActionRow,
    formActions,
    formGrid,
    formRow,
    headerActions,
    inlineActions,
    input,
    loginField,
    loginInput,
    loginPasswordInput,
    listItem,
    listMeta,
    listRow,
    listTitle,
    luaCodeBlock,
    codeToolbar,
    paneMeta,
    passwordInput,
    pill,
    problemSection,
    section,
    stack,
    textarea
  };

  Object.assign(window.OJLite, {
    bindResizablePane,
    createColumnResizer,
    latestVerdictLabel,
    latestVerdictPillClass,
    statusPillClass,
    verdictLabel,
    verdictPillClass,
    ui
  });
})();
