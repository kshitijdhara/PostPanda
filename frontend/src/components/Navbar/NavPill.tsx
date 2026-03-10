import { Link, useLocation } from 'react-router-dom';
import { PencilLine, BookOpen02 } from '@untitled-ui/icons-react';
import styles from './NavPill.module.scss';

export default function NavPill() {
  const location = useLocation();
  const isWrite = location.pathname === '/posts/new';
  const isDrafts = location.pathname === '/drafts';
  const active = isWrite ? 'write' : isDrafts ? 'drafts' : 'write';

  return (
    <div className={styles.pillContainer}>
      <div className={styles.pillTrack}>
        <div className={`${styles.pillSlider} ${styles[`slider-${active}`]}`} />

        <Link
          to="/posts/new"
          className={`${styles.pillButton} ${active === 'write' ? styles.active : ''}`}
        >
          <PencilLine width={18} height={18} />
          <span>Write</span>
        </Link>

        <Link
          to="/drafts"
          className={`${styles.pillButton} ${active === 'drafts' ? styles.active : ''}`}
        >
          <BookOpen02 width={18} height={18} />
          <span>Drafts</span>
        </Link>
      </div>
    </div>
  );
}
