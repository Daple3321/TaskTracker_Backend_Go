import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { apiFetch } from "../api/client";
import { useAuth } from "../context/AuthContext";
import { insertCodeBlock, insertLink, prefixHeading, wrapSelection } from "../utils/markdownSelection";

type Tag = {
  id: number;
  userId: number;
  name: string;
};

type Task = {
  id: number;
  name: string;
  description: string;
  userId: number;
  createdAt: string;
  updatedAt: string;
  tags: Tag[];
};

type TasksResponse = {
  tasks: Task[];
  page: number;
  limit: number;
  total_items: number;
  total_pages: number;
};

function TrashIcon({ onClick }: { onClick: (e: React.MouseEvent) => void }) {
  return (
    <button
      type="button"
      className="icon-btn danger"
      onClick={onClick}
      onMouseDown={(e) => e.preventDefault()}
      aria-label="Delete task"
    >
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <polyline points="3 6 5 6 21 6" />
        <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
        <line x1="10" y1="11" x2="10" y2="17" />
        <line x1="14" y1="11" x2="14" y2="17" />
      </svg>
    </button>
  );
}

function normalizeTask(task: Task): Task {
  return {
    ...task,
    tags: Array.isArray(task.tags) ? task.tags : [],
  };
}

function sortTags(tags: Tag[]): Tag[] {
  return [...tags].sort((a, b) => a.name.localeCompare(b.name));
}

function summarizeDescription(markdown: string): string {
  const plainText = markdown
    .replace(/```[\s\S]*?```/g, " ")
    .replace(/`[^`]*`/g, " ")
    .replace(/!\[[^\]]*]\([^)]*\)/g, " ")
    .replace(/\[[^\]]*]\([^)]*\)/g, " ")
    .replace(/[#>*_~|-]/g, " ")
    .replace(/\s+/g, " ")
    .trim();

  if (plainText.length <= 80) return plainText;
  return `${plainText.slice(0, 77).trim()}…`;
}

function formatTs(value: string): string {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

export function Dashboard() {
  const { user, token, logout } = useAuth();
  const navigate = useNavigate();
  const [tasks, setTasks] = useState<Task[]>([]);
  const [tasksError, setTasksError] = useState("");
  const [tasksLoading, setTasksLoading] = useState(true);
  const [tags, setTags] = useState<Tag[]>([]);
  const [tagsError, setTagsError] = useState("");
  const [tagsLoading, setTagsLoading] = useState(true);
  const [selectedTagId, setSelectedTagId] = useState<number | null>(null);
  const [newTaskOpen, setNewTaskOpen] = useState(false);
  const [newTaskTitle, setNewTaskTitle] = useState("");
  const [newTaskSubmitting, setNewTaskSubmitting] = useState(false);
  const [globalTagInputOpen, setGlobalTagInputOpen] = useState(false);
  const [globalTagDraft, setGlobalTagDraft] = useState("");
  const [globalTagBusy, setGlobalTagBusy] = useState(false);
  const [selectedTaskId, setSelectedTaskId] = useState<number | null>(null);
  const [panelName, setPanelName] = useState("");
  const [panelDesc, setPanelDesc] = useState("");
  const [panelError, setPanelError] = useState("");
  const [panelSaving, setPanelSaving] = useState(false);
  const [descEditing, setDescEditing] = useState(false);
  const [descDraft, setDescDraft] = useState("");
  const [taskTagMenuOpen, setTaskTagMenuOpen] = useState(false);
  const [newTaskTagName, setNewTaskTagName] = useState("");
  const [taskTagBusy, setTaskTagBusy] = useState(false);
  const taskTagMenuRef = useRef<HTMLDivElement>(null);
  const descEditorRef = useRef<HTMLTextAreaElement>(null);
  const mdToolbarRef = useRef<HTMLDivElement>(null);

  const applyDescEdit = useCallback((fn: (value: string, start: number, end: number) => ReturnType<typeof wrapSelection>) => {
    const el = descEditorRef.current;
    if (!el) return;
    const start = el.selectionStart;
    const end = el.selectionEnd;
    setDescDraft((prev) => {
      const { next, selStart, selEnd } = fn(prev, start, end);
      setTimeout(() => {
        el.focus();
        el.setSelectionRange(selStart, selEnd);
      }, 0);
      return next;
    });
  }, []);

  const loadTasks = useCallback(async () => {
    if (!token) return;
    try {
      const res = await apiFetch("/tasks/?page=1&limit=50", { token });
      if (!res.ok) {
        setTasksError("Could not load tasks");
        return;
      }
      const data: TasksResponse = await res.json();
      setTasks((data.tasks ?? []).map(normalizeTask));
      setTasksError("");
    } catch {
      setTasksError("Could not load tasks");
    } finally {
      setTasksLoading(false);
    }
  }, [token]);

  const loadTags = useCallback(async () => {
    if (!token) return;
    try {
      const res = await apiFetch("/tasks/tags/", { token });
      if (!res.ok) {
        setTagsError("Could not load tags");
        return;
      }
      const data: Tag[] = await res.json();
      setTags(sortTags(data ?? []));
      setTagsError("");
    } catch {
      setTagsError("Could not load tags");
    } finally {
      setTagsLoading(false);
    }
  }, [token]);

  useEffect(() => {
    if (!token) return;
    let cancelled = false;
    setTasksLoading(true);
    setTagsLoading(true);
    (async () => {
      await Promise.all([loadTasks(), loadTags()]);
      if (!cancelled) {
        setTasksLoading(false);
        setTagsLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [token, loadTasks, loadTags]);

  const selectedTask = useMemo(
    () => tasks.find((task) => task.id === selectedTaskId) ?? null,
    [tasks, selectedTaskId]
  );

  const visibleTasks = useMemo(() => {
    if (selectedTagId == null) return tasks;
    return tasks.filter((task) => task.tags.some((tag) => tag.id === selectedTagId));
  }, [tasks, selectedTagId]);

  const tagsNotOnTask = useMemo(() => {
    if (!selectedTask) return [];
    const onIds = new Set(selectedTask.tags.map((t) => t.id));
    return tags.filter((t) => !onIds.has(t.id));
  }, [selectedTask, tags]);

  useEffect(() => {
    if (!selectedTaskId) return;
    if (!tasks.some((t) => t.id === selectedTaskId)) {
      setSelectedTaskId(null);
      setPanelError("");
    }
  }, [tasks, selectedTaskId]);

  useEffect(() => {
    function onDocMouseDown(e: MouseEvent) {
      if (!taskTagMenuOpen) return;
      const el = taskTagMenuRef.current;
      if (el && !el.contains(e.target as Node)) {
        setTaskTagMenuOpen(false);
        setNewTaskTagName("");
      }
    }
    document.addEventListener("mousedown", onDocMouseDown);
    return () => document.removeEventListener("mousedown", onDocMouseDown);
  }, [taskTagMenuOpen]);

  function handleLogout() {
    logout();
    navigate("/login", { replace: true });
  }

  function selectTask(task: Task) {
    setSelectedTaskId(task.id);
    setPanelName(task.name);
    setPanelDesc(task.description ?? "");
    setPanelError("");
    setDescEditing(false);
    setDescDraft(task.description ?? "");
    setTaskTagMenuOpen(false);
    setNewTaskTagName("");
  }

  function toggleTagFilter(tagId: number) {
    setSelectedTagId((current) => (current === tagId ? null : tagId));
  }

  const handleSaveTask = useCallback(
    async (descriptionOverride?: string) => {
      if (!token || !selectedTask) return;

      const nextName = panelName.trim();
      const nextDesc = (descriptionOverride ?? panelDesc).trim();
      if (!nextName) {
        setPanelError("Title required");
        return;
      }

      setPanelError("");
      setPanelSaving(true);
      try {
        const res = await apiFetch(`/tasks/${selectedTask.id}/`, {
          method: "PUT",
          token,
          body: JSON.stringify({ name: nextName, description: nextDesc }),
        });
        if (!res.ok) {
          const text = await res.text();
          setPanelError(text || "Could not save");
          return;
        }
        const updated: Task = normalizeTask(await res.json());
        setTasks((prev) => prev.map((t) => (t.id === selectedTask.id ? updated : t)));
        setPanelName(updated.name);
        setPanelDesc(updated.description);
        setDescDraft(updated.description);
      } catch {
        setPanelError("Could not save");
      } finally {
        setPanelSaving(false);
      }
    },
    [panelDesc, panelName, selectedTask, token]
  );

  async function createGlobalTag() {
    if (!token) return;
    const name = globalTagDraft.trim();
    if (!name) return;

    setGlobalTagBusy(true);
    try {
      const res = await apiFetch("/tasks/tags/", {
        method: "POST",
        token,
        body: JSON.stringify({ name }),
      });
      if (!res.ok) {
        const text = await res.text();
        setTagsError(text || "Could not create tag");
        return;
      }
      const created: Tag = await res.json();
      setTags((prev) => sortTags([...prev, created]));
      setGlobalTagDraft("");
      setGlobalTagInputOpen(false);
      setTagsError("");
    } catch {
      setTagsError("Could not create tag");
    } finally {
      setGlobalTagBusy(false);
    }
  }

  async function deleteGlobalTag(tagId: number) {
    if (!token) return;
    try {
      const res = await apiFetch(`/tasks/tags/${tagId}/`, { method: "DELETE", token });
      if (!res.ok) {
        const text = await res.text();
        setTagsError(text || "Could not delete tag");
        return;
      }
      setTags((prev) => prev.filter((t) => t.id !== tagId));
      setTasks((prev) =>
        prev.map((task) => ({
          ...task,
          tags: task.tags.filter((t) => t.id !== tagId),
        }))
      );
      if (selectedTagId === tagId) setSelectedTagId(null);
      setTagsError("");
    } catch {
      setTagsError("Could not delete tag");
    }
  }

  async function handleNewTask(e: React.FormEvent) {
    e.preventDefault();
    if (!token) return;
    const title = newTaskTitle.trim();
    if (!title) return;

    setNewTaskSubmitting(true);
    try {
      const res = await apiFetch("/tasks/", {
        method: "POST",
        token,
        body: JSON.stringify({ name: title, description: "" }),
      });
      if (!res.ok) {
        const text = await res.text();
        setTasksError(text || "Could not create task");
        return;
      }
      const created = normalizeTask(await res.json());
      setTasks((prev) => [created, ...prev]);
      setNewTaskTitle("");
      setNewTaskOpen(false);
      selectTask(created);
      setTasksError("");
    } catch {
      setTasksError("Could not create task");
    } finally {
      setNewTaskSubmitting(false);
    }
  }

  const handleDeleteTask = useCallback(
    async (id: number) => {
      if (!token) return;
      try {
        const res = await apiFetch(`/tasks/${id}/`, { method: "DELETE", token });
        if (!res.ok) return;
        setTasks((prev) => prev.filter((t) => t.id !== id));
        if (selectedTaskId === id) {
          setSelectedTaskId(null);
          setDescEditing(false);
        }
      } catch {
        // ignore
      }
    },
    [selectedTaskId, token]
  );

  async function attachTagToTask(tag: Tag) {
    if (!token || !selectedTask) return;
    setTaskTagBusy(true);
    setPanelError("");
    try {
      const res = await apiFetch(`/tasks/task/${selectedTask.id}/tags/`, {
        method: "POST",
        token,
        body: JSON.stringify({ name: tag.name }),
      });
      if (!res.ok) {
        const text = await res.text();
        setPanelError(text || "Could not add tag");
        return;
      }
      setTaskTagMenuOpen(false);
      setNewTaskTagName("");
      await Promise.all([loadTasks(), loadTags()]);
    } catch {
      setPanelError("Could not add tag");
    } finally {
      setTaskTagBusy(false);
    }
  }

  async function createAndAttachTag() {
    if (!token || !selectedTask) return;
    const name = newTaskTagName.trim();
    if (!name) return;

    setTaskTagBusy(true);
    setPanelError("");
    try {
      const res = await apiFetch(`/tasks/task/${selectedTask.id}/tags/`, {
        method: "POST",
        token,
        body: JSON.stringify({ name }),
      });
      if (!res.ok) {
        const text = await res.text();
        setPanelError(text || "Could not add tag");
        return;
      }
      setNewTaskTagName("");
      setTaskTagMenuOpen(false);
      await Promise.all([loadTasks(), loadTags()]);
    } catch {
      setPanelError("Could not add tag");
    } finally {
      setTaskTagBusy(false);
    }
  }

  async function detachTag(tagId: number) {
    if (!token || !selectedTask) return;
    setPanelError("");
    try {
      const res = await apiFetch(`/tasks/task/${selectedTask.id}/tags/${tagId}/`, {
        method: "DELETE",
        token,
      });
      if (!res.ok) {
        const text = await res.text();
        setPanelError(text || "Could not remove tag");
        return;
      }
      setTasks((prev) =>
        prev.map((task) =>
          task.id === selectedTask.id
            ? { ...task, tags: task.tags.filter((t) => t.id !== tagId) }
            : task
        )
      );
    } catch {
      setPanelError("Could not remove tag");
    }
  }

  function startDescEdit() {
    if (!selectedTask) return;
    setDescDraft(panelDesc);
    setDescEditing(true);
  }

  async function finishDescEdit(save: boolean) {
    if (!selectedTask) {
      setDescEditing(false);
      return;
    }
    if (!save) {
      setPanelDesc(selectedTask.description ?? "");
      setDescDraft(selectedTask.description ?? "");
      setDescEditing(false);
      return;
    }
    const next = descDraft;
    setPanelDesc(next);
    setDescEditing(false);
    await handleSaveTask(next);
  }

  function onDescKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === "Escape") {
      e.preventDefault();
      void finishDescEdit(false);
    }
  }

  function onDescBlur(e: React.FocusEvent<HTMLTextAreaElement>) {
    const rel = e.relatedTarget as Node | null;
    if (mdToolbarRef.current && rel && mdToolbarRef.current.contains(rel)) {
      return;
    }
    void finishDescEdit(true);
  }

  return (
    <div className="dashboard-workbench">
      <aside className="sidebar-rail">
        <div className="sidebar-header">
          <div>
            <div className="sidebar-greeting">Hello, {user?.username ?? "—"}</div>
          </div>
          <button type="button" className="btn-ghost" onClick={handleLogout}>
            Log out
          </button>
        </div>

        <div className="sidebar-tags">
          {(tagsError || tasksError) && (
            <p className="error sidebar-error">{tagsError || tasksError}</p>
          )}
          <div className="tag-filter-bar sidebar-tag-chips">
            <button
              type="button"
              className={`tag-chip tag-chip-filter${selectedTagId == null ? " is-active" : ""}`}
              onClick={() => setSelectedTagId(null)}
            >
              All
            </button>
            {tagsLoading && <span className="muted tiny">…</span>}
            {tags.map((tag) => (
              <div
                key={tag.id}
                className={`tag-chip tag-chip-manage${selectedTagId === tag.id ? " is-active" : ""}`}
              >
                <button type="button" className="tag-chip-label" onClick={() => toggleTagFilter(tag.id)}>
                  {tag.name}
                </button>
                <button
                  type="button"
                  className="tag-chip-remove"
                  aria-label={`Delete ${tag.name}`}
                  onClick={() => deleteGlobalTag(tag.id)}
                >
                  ×
                </button>
              </div>
            ))}
            {globalTagInputOpen ? (
              <input
                className="sidebar-tag-inline-input"
                value={globalTagDraft}
                onChange={(e) => setGlobalTagDraft(e.target.value)}
                placeholder="Name"
                autoFocus
                disabled={globalTagBusy}
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    e.preventDefault();
                    void createGlobalTag();
                  }
                  if (e.key === "Escape") {
                    setGlobalTagInputOpen(false);
                    setGlobalTagDraft("");
                  }
                }}
                onBlur={() => {
                  if (!globalTagDraft.trim()) setGlobalTagInputOpen(false);
                }}
              />
            ) : (
              <button
                type="button"
                className="tag-chip tag-add-plus"
                aria-label="New tag"
                onClick={() => setGlobalTagInputOpen(true)}
              >
                +
              </button>
            )}
          </div>
        </div>

        <div className="sidebar-list-wrap">
          {tasksLoading && <p className="muted tiny pad">Loading…</p>}
          {!tasksLoading && visibleTasks.length === 0 && (
            <p className="muted tiny pad">{selectedTagId == null ? "No tasks" : "No match"}</p>
          )}
          <ul className="task-list sidebar-task-list">
            {visibleTasks.map((task) => {
              const summary = summarizeDescription(task.description ?? "");
              return (
                <li key={task.id}>
                  <button
                    type="button"
                    className={`sidebar-task-item${selectedTaskId === task.id ? " is-selected" : ""}`}
                    onClick={() => selectTask(task)}
                  >
                    <span className="sidebar-task-title">{task.name}</span>
                    {task.tags.length > 0 && (
                      <span className="sidebar-task-tags">
                        {task.tags.map((t) => t.name).join(", ")}
                      </span>
                    )}
                    {summary && <span className="sidebar-task-preview">{summary}</span>}
                  </button>
                </li>
              );
            })}
          </ul>
        </div>

        <div className="sidebar-footer">
          {newTaskOpen ? (
            <form className="sidebar-new-task" onSubmit={handleNewTask}>
              <input
                value={newTaskTitle}
                onChange={(e) => setNewTaskTitle(e.target.value)}
                placeholder="Title"
                autoFocus
                disabled={newTaskSubmitting}
              />
              <div className="sidebar-new-task-actions">
                <button type="submit" className="btn-primary btn-tiny" disabled={newTaskSubmitting}>
                  Add
                </button>
                <button
                  type="button"
                  className="btn-ghost btn-tiny"
                  onClick={() => {
                    setNewTaskOpen(false);
                    setNewTaskTitle("");
                  }}
                >
                  Cancel
                </button>
              </div>
            </form>
          ) : (
            <button
              type="button"
              className="sidebar-add-task"
              aria-label="New task"
              onClick={() => setNewTaskOpen(true)}
            >
              +
            </button>
          )}
        </div>
      </aside>

      <main className="main-pane">
        {!selectedTask ? (
          <div className="main-empty">
            <p className="muted">Select a task</p>
          </div>
        ) : (
          <>
            <div className="main-pane-header">
              <input
                className="main-title-input"
                value={panelName}
                onChange={(e) => setPanelName(e.target.value)}
                aria-label="Task title"
              />
              <div className="main-pane-header-actions">
                <button
                  type="button"
                  className="btn-primary main-save-btn"
                  disabled={panelSaving}
                  onClick={() => void handleSaveTask()}
                >
                  {panelSaving ? "…" : "Save"}
                </button>
                <TrashIcon onClick={() => void handleDeleteTask(selectedTask.id)} />
              </div>
            </div>
            <p className="main-meta muted tiny">
              Created {formatTs(selectedTask.createdAt) || "—"} · Updated {formatTs(selectedTask.updatedAt) || "—"}
            </p>

            {panelError && <p className="error">{panelError}</p>}

            <div className="main-tags-row" ref={taskTagMenuRef}>
              {selectedTask.tags.map((tag) => (
                <div key={tag.id} className="tag-chip tag-chip-manage">
                  <span className="tag-chip-label-static">{tag.name}</span>
                  <button
                    type="button"
                    className="tag-chip-remove"
                    aria-label={`Remove ${tag.name}`}
                    onClick={() => void detachTag(tag.id)}
                  >
                    ×
                  </button>
                </div>
              ))}
              <div className="tag-add-wrap">
                <button
                  type="button"
                  className="tag-chip tag-add-plus"
                  aria-label="Add tag"
                  onClick={() => setTaskTagMenuOpen((o) => !o)}
                >
                  +
                </button>
                {taskTagMenuOpen && (
                  <div className="tag-add-popover">
                    <div className="tag-add-popover-title">Tags</div>
                    {tagsNotOnTask.length === 0 ? (
                      <p className="muted tiny popover-hint">All tags attached</p>
                    ) : (
                      <ul className="tag-add-list">
                        {tagsNotOnTask.map((tag) => (
                          <li key={tag.id}>
                            <button
                              type="button"
                              className="tag-add-list-btn"
                              disabled={taskTagBusy}
                              onClick={() => void attachTagToTask(tag)}
                            >
                              {tag.name}
                            </button>
                          </li>
                        ))}
                      </ul>
                    )}
                    <div className="tag-add-new-row">
                      <input
                        value={newTaskTagName}
                        onChange={(e) => setNewTaskTagName(e.target.value)}
                        placeholder="New tag name"
                        disabled={taskTagBusy}
                        onKeyDown={(e) => {
                          if (e.key === "Enter") {
                            e.preventDefault();
                            void createAndAttachTag();
                          }
                        }}
                      />
                    </div>
                  </div>
                )}
              </div>
            </div>

            <div className="main-md-shell">
              {descEditing ? (
                <>
                  <div
                    ref={mdToolbarRef}
                    className="main-md-toolbar"
                    role="toolbar"
                    aria-label="Markdown formatting"
                  >
                    <button
                      type="button"
                      title="Bold"
                      onMouseDown={(e) => e.preventDefault()}
                      onClick={() => applyDescEdit((v, s, e) => wrapSelection(v, s, e, "**", "**"))}
                    >
                      Bold
                    </button>
                    <button
                      type="button"
                      title="Italic"
                      onMouseDown={(e) => e.preventDefault()}
                      onClick={() => applyDescEdit((v, s, e) => wrapSelection(v, s, e, "*", "*"))}
                    >
                      Italic
                    </button>
                    <button
                      type="button"
                      title="Heading level 1"
                      onMouseDown={(e) => e.preventDefault()}
                      onClick={() => applyDescEdit((v, s, e) => prefixHeading(v, Math.min(s, e), "#"))}
                    >
                      H1
                    </button>
                    <button
                      type="button"
                      title="Heading level 2"
                      onMouseDown={(e) => e.preventDefault()}
                      onClick={() => applyDescEdit((v, s, e) => prefixHeading(v, Math.min(s, e), "##"))}
                    >
                      H2
                    </button>
                    <button
                      type="button"
                      title="Heading level 3"
                      onMouseDown={(e) => e.preventDefault()}
                      onClick={() => applyDescEdit((v, s, e) => prefixHeading(v, Math.min(s, e), "###"))}
                    >
                      H3
                    </button>
                    <button
                      type="button"
                      title="Link"
                      onMouseDown={(e) => e.preventDefault()}
                      onClick={() => applyDescEdit((v, s, e) => insertLink(v, s, e))}
                    >
                      Link
                    </button>
                    <button
                      type="button"
                      title="Inline code"
                      onMouseDown={(e) => e.preventDefault()}
                      onClick={() => applyDescEdit((v, s, e) => wrapSelection(v, s, e, "`", "`"))}
                    >
                      Code
                    </button>
                    <button
                      type="button"
                      title="Code block"
                      onMouseDown={(e) => e.preventDefault()}
                      onClick={() => applyDescEdit((v, s, e) => insertCodeBlock(v, s, e))}
                    >
                      Block
                    </button>
                  </div>
                  <textarea
                    ref={descEditorRef}
                    className="main-md-editor"
                    value={descDraft}
                    onChange={(e) => setDescDraft(e.target.value)}
                    onKeyDown={onDescKeyDown}
                    onBlur={onDescBlur}
                    autoFocus
                    spellCheck
                  />
                </>
              ) : (
                <div
                  className="main-md-preview markdown-preview"
                  role="button"
                  tabIndex={0}
                  onDoubleClick={startDescEdit}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" || e.key === " ") {
                      e.preventDefault();
                      startDescEdit();
                    }
                  }}
                >
                  {panelDesc.trim() ? (
                    <ReactMarkdown remarkPlugins={[remarkGfm]}>{panelDesc}</ReactMarkdown>
                  ) : (
                    <p className="muted md-placeholder">Double-click to edit</p>
                  )}
                </div>
              )}
            </div>
          </>
        )}
      </main>
    </div>
  );
}
