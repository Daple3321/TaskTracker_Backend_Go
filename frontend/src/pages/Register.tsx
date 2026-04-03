import { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { apiFetch } from "../api/client";

export function Register() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      const res = await apiFetch("/auth/register", {
        method: "POST",
        body: JSON.stringify({ username, password }),
      });
      if (!res.ok) {
        const text = await res.text();
        setError(text || res.statusText || "Registration failed");
        return;
      }
      navigate("/login", { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Request failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="page card">
      <h1>Register</h1>
      <form onSubmit={handleSubmit}>
        {error && <p className="error">{error}</p>}
        <div className="field">
          <label htmlFor="reg-username">Username</label>
          <input
            id="reg-username"
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            required
            autoComplete="username"
          />
        </div>
        <div className="field">
          <label htmlFor="reg-password">Password</label>
          <input
            id="reg-password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            autoComplete="new-password"
          />
        </div>
        <button type="submit" disabled={loading}>
          {loading ? "Creating account…" : "Create account"}
        </button>
      </form>
      <p className="link-line">
        Already have an account? <Link to="/login">Log in</Link>
      </p>
    </div>
  );
}
