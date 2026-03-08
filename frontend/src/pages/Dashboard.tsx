import { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiFetch } from "../api/client";
import { useAuth } from "../context/AuthContext";

type Task = {
  id: number;
  name: string;
  description: string;
  userId: number;
  createdAt: string;
  updatedAt: string;
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
      className="task-delete"
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

export function Dashboard() {
  const { user, token, logout } = useAuth();
  const navigate = useNavigate();
  const [tasks, setTasks] = useState<Task[]>([]);
  const [tasksError, setTasksError] = useState("");
  const [tasksLoading, setTasksLoading] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);
  const [createName, setCreateName] = useState("");
  const [createDesc, setCreateDesc] = useState("");
  const [createError, setCreateError] = useState("");
  const [createSubmitting, setCreateSubmitting] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editName, setEditName] = useState("");
  const [editDesc, setEditDesc] = useState("");
  const [editError, setEditError] = useState("");
  const editWrapperRef = useRef<HTMLDivElement>(null);

  const loadTasks = useCallback(async () => {
    if (!token) return;
    try {
      const res = await apiFetch("/tasks/?page=1&limit=50", { token });
      if (!res.ok) {
        setTasksError("Could not load tasks");
        return;
      }
      const data: TasksResponse = await res.json();
      setTasks(data.tasks ?? []);
      setTasksError("");
    } catch {
      setTasksError("Could not load tasks");
    } finally {
      setTasksLoading(false);
    }
  }, [token]);

  useEffect(() => {
    if (!token) return;
    let cancelled = false;
    setTasksLoading(true);
    (async () => {
      await loadTasks();
      if (!cancelled) setTasksLoading(false);
    })();
    return () => { cancelled = true; };
  }, [token, loadTasks]);

  function handleLogout() {
    logout();
    navigate("/login", { replace: true });
  }

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setCreateError("");
    setCreateSubmitting(true);
    try {
      const res = await apiFetch("/tasks/", {
        method: "POST",
        token,
        body: JSON.stringify({ name: createName.trim(), description: createDesc.trim() }),
      });
      if (!res.ok) {
        const text = await res.text();
        setCreateError(text || "Could not create task");
        return;
      }
      const created: Task = await res.json();
      setTasks((prev) => [created, ...prev]);
      setCreateName("");
      setCreateDesc("");
      setCreateOpen(false);
    } catch {
      setCreateError("Could not create task");
    } finally {
      setCreateSubmitting(false);
    }
  }

  function startEdit(t: Task) {
    setEditingId(t.id);
    setEditName(t.name);
    setEditDesc(t.description);
    setEditError("");
  }

  const submitEdit = useCallback(async (id: number) => {
    if (!token) return;
    setEditError("");
    try {
      const res = await apiFetch(`/tasks/${id}/`, {
        method: "PUT",
        token,
        body: JSON.stringify({ name: editName.trim(), description: editDesc.trim() }),
      });
      if (!res.ok) {
        const text = await res.text();
        setEditError(text || "Could not update task");
        return;
      }
      const updated: Task = await res.json();
      setTasks((prev) => prev.map((x) => (x.id === id ? updated : x)));
      setEditingId(null);
    } catch {
      setEditError("Could not update task");
    }
  }, [token, editName, editDesc]);

  function handleEditKeyDown(e: React.KeyboardEvent, id: number) {
    if (e.key === "Enter") {
      e.preventDefault();
      submitEdit(id);
    }
    if (e.key === "Escape") {
      setEditingId(null);
      setEditError("");
    }
  }

  function handleEditBlur(id: number) {
    if (editingId !== id) return;
    setTimeout(() => {
      const wrapper = editWrapperRef.current;
      if (wrapper?.contains(document.activeElement)) return;
      submitEdit(id);
    }, 0);
  }

  async function handleDelete(id: number) {
    if (!token) return;
    try {
      const res = await apiFetch(`/tasks/${id}/`, { method: "DELETE", token });
      if (!res.ok) return;
      setTasks((prev) => prev.filter((t) => t.id !== id));
      if (editingId === id) setEditingId(null);
    } catch {
      // ignore
    }
  }

  return (
    <div className="page card">
      <div className="dashboard-header">
        <h1>Hello, {user?.username ?? "—"}.</h1>
        <button type="button" onClick={handleLogout}>
          Log out
        </button>
      </div>

      <section className="tasks-section">
        <div className="tasks-section-header">
          <h2>Your tasks</h2>
          <button
            type="button"
            className="btn-primary"
            onClick={() => {
              setCreateOpen(true);
              setCreateError("");
            }}
          >
            New task
          </button>
        </div>

        {createOpen && (
          <form className="task-create-form" onSubmit={handleCreate}>
            {createError && <p className="error">{createError}</p>}
            <div className="field">
              <label htmlFor="create-name">Title</label>
              <input
                id="create-name"
                value={createName}
                onChange={(e) => setCreateName(e.target.value)}
                placeholder="Task title"
                required
              />
            </div>
            <div className="field">
              <label htmlFor="create-desc">Description</label>
              <input
                id="create-desc"
                value={createDesc}
                onChange={(e) => setCreateDesc(e.target.value)}
                placeholder="Optional description"
              />
            </div>
            <div className="form-actions">
              <button type="submit" className="btn-primary" disabled={createSubmitting}>
                {createSubmitting ? "Adding…" : "Add"}
              </button>
              <button
                type="button"
                onClick={() => {
                  setCreateOpen(false);
                  setCreateError("");
                }}
              >
                Cancel
              </button>
            </div>
          </form>
        )}

        {tasksLoading && <p>Loading tasks…</p>}
        {tasksError && <p className="error">{tasksError}</p>}
        {!tasksLoading && !tasksError && tasks.length === 0 && !createOpen && (
          <p className="muted">No tasks yet.</p>
        )}
        {!tasksLoading && tasks.length > 0 && (
          <ul className="task-list">
            {tasks.map((t) => (
              <li
                key={t.id}
                className={`task-row${editingId === t.id ? " task-editing" : ""}`}
              >
                {editingId === t.id ? (
                  <>
                    <div
                      ref={editWrapperRef}
                      className="task-edit"
                      onBlur={() => handleEditBlur(t.id)}
                    >
                      {editError && <p className="error task-edit-error">{editError}</p>}
                      <input
                        className="task-edit-title"
                        value={editName}
                        onChange={(e) => setEditName(e.target.value)}
                        onKeyDown={(e) => handleEditKeyDown(e, t.id)}
                        placeholder="Title"
                        autoFocus
                      />
                      <input
                        className="task-edit-desc"
                        value={editDesc}
                        onChange={(e) => setEditDesc(e.target.value)}
                        onKeyDown={(e) => handleEditKeyDown(e, t.id)}
                        placeholder="Description"
                      />
                      <p className="task-edit-hint">Enter to save, Escape to cancel</p>
                    </div>
                    <TrashIcon
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDelete(t.id);
                      }}
                    />
                  </>
                ) : (
                  <>
                    <div
                      className="task-content"
                      onClick={() => startEdit(t)}
                      role="button"
                      tabIndex={0}
                      onKeyDown={(e) => {
                        if (e.key === "Enter" || e.key === " ") {
                          e.preventDefault();
                          startEdit(t);
                        }
                      }}
                    >
                      <strong>{t.name}</strong>
                      {t.description ? ` — ${t.description}` : ""}
                    </div>
                    <TrashIcon
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDelete(t.id);
                      }}
                    />
                  </>
                )}
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}
