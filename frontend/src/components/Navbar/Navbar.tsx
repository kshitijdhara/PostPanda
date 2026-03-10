import { useState, useRef, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { LogOut01, LogIn01, User01 } from '@untitled-ui/icons-react';
import { useAuth } from '../../features/auth/AuthContext';
import NavPill from './NavPill';
import styles from './Navbar.module.scss';

function getInitial(displayName: string): string {
  return displayName.charAt(0).toUpperCase();
}

function getBannerGradient(username: string): string {
  let hash = 0;
  for (let i = 0; i < username.length; i++) {
    hash = username.charCodeAt(i) + ((hash << 5) - hash);
  }
  const h1 = Math.abs(hash) % 360;
  const h2 = (h1 + 40) % 360;
  return `linear-gradient(135deg, hsl(${h1}, 60%, 55%), hsl(${h2}, 70%, 45%))`;
}

export default function Navbar() {
  const { user, isAuthenticated, logout } = useAuth();
  const navigate = useNavigate();
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const handleLogout = async () => {
    await logout();
    navigate('/');
  };

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setDropdownOpen(false);
      }
    }

    if (dropdownOpen) {
      document.addEventListener('mousedown', handleClickOutside);
      return () => document.removeEventListener('mousedown', handleClickOutside);
    }
  }, [dropdownOpen]);

  return (
    <nav className={styles.navbar}>
      <div className={styles.inner}>
        <Link to="/" className={styles.logo}>📝 Saga</Link>
        <div className={styles.nav}>
          {isAuthenticated ? (
            <>
              <NavPill />

              {/* Profile dropdown */}
              <div className={styles.profileDropdown} ref={dropdownRef}>
                <button
                  className={styles.profileButton}
                  onClick={() => setDropdownOpen(!dropdownOpen)}
                  aria-label="User menu"
                  aria-expanded={dropdownOpen}
                >
                  {user?.avatar_url ? (
                    <img src={user.avatar_url} alt={user.display_name} className={styles.profileImage} />
                  ) : (
                    <div
                      className={styles.profileImageFallback}
                      style={{ background: getBannerGradient(user?.username || '') }}
                    >
                      {getInitial(user?.display_name || '')}
                    </div>
                  )}
                </button>

                {dropdownOpen && (
                  <div className={styles.dropdownMenu}>
                    <div className={styles.dropdownHeader}>
                      <p className={styles.dropdownName}>{user?.display_name}</p>
                      <p className={styles.dropdownUsername}>@{user?.username}</p>
                    </div>
                    <div className={styles.dropdownDivider} />
                    <Link
                      to={`/profile/${user?.username}`}
                      className={styles.dropdownItem}
                      onClick={() => setDropdownOpen(false)}
                    >
                      <User01 width={16} height={16} />
                      <span>View Profile</span>
                    </Link>
                    <button
                      className={styles.dropdownItem + ' ' + styles.logoutItem}
                      onClick={() => {
                        setDropdownOpen(false);
                        handleLogout();
                      }}
                    >
                      <LogOut01 width={16} height={16} />
                      <span>Logout</span>
                    </button>
                  </div>
                )}
              </div>
            </>
          ) : (
            <>
              <Link to="/login" className={styles.link}>
                <LogIn01 width={18} height={18} />
                <span>Login</span>
              </Link>
              <Link to="/register" className={styles.writeBtn}>
                <User01 width={18} height={18} />
                <span>Get Started</span>
              </Link>
            </>
          )}
        </div>
      </div>
    </nav>
  );
}
