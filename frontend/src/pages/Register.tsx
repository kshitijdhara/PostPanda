import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../features/auth/AuthContext';
import styles from './Auth.module.scss';

export default function Register() {
  const { register } = useAuth();
  const navigate = useNavigate();
  const [form, setForm] = useState({
    username: '',
    email: '',
    password: '',
    display_name: '',
  });
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const update = (key: string) => (e: React.ChangeEvent<HTMLInputElement>) =>
    setForm((prev) => ({ ...prev, [key]: e.target.value }));

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    try {
      await register(form);
      navigate('/');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className={styles.container}>
      <h1 className={styles.title}>Create an account</h1>
      <p className={styles.subtitle}>Join the community and start writing</p>

      <form onSubmit={handleSubmit} className={styles.form}>
        {error && <div className={styles.error}>{error}</div>}

        <div className={styles.field}>
          <label className={styles.label}>Display Name</label>
          <input type="text" className={styles.input} value={form.display_name} onChange={update('display_name')} required />
        </div>

        <div className={styles.field}>
          <label className={styles.label}>Username</label>
          <input type="text" className={styles.input} value={form.username} onChange={update('username')} required />
        </div>

        <div className={styles.field}>
          <label className={styles.label}>Email</label>
          <input type="email" className={styles.input} value={form.email} onChange={update('email')} required />
        </div>

        <div className={styles.field}>
          <label className={styles.label}>Password</label>
          <input type="password" className={styles.input} value={form.password} onChange={update('password')} required minLength={6} />
        </div>

        <button type="submit" className={styles.submitBtn} disabled={isLoading}>
          {isLoading ? 'Creating account...' : 'Create account'}
        </button>
      </form>

      <p className={styles.footer}>
        Already have an account? <Link to="/login">Sign in</Link>
      </p>
    </div>
  );
}
