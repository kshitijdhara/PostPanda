import { useParams, Link, useNavigate } from 'react-router-dom';
import {
  useUserProfile,
  useUserPosts,
  useUserComments,
  useMyLikedPosts,
  useMyComments,
  useMyBookmarkedPosts,
  useDrafts,
  useUpdateProfile,
  useChangePassword,
} from '../hooks/useProfile';
import { useAuth } from '../features/auth/AuthContext';
import PostCard from '../components/PostCard/PostCard';
import { ApiError } from '../services/api';
import styles from './Profile.module.scss';
import { useState } from 'react';

type Tab = 'posts' | 'comments' | 'liked' | 'bookmarks' | 'drafts';

function getBannerGradient(username: string): string {
  let hash = 0;
  for (let i = 0; i < username.length; i++) {
    hash = username.charCodeAt(i) + ((hash << 5) - hash);
  }
  const h1 = Math.abs(hash) % 360;
  const h2 = (h1 + 40) % 360;
  return `linear-gradient(135deg, hsl(${h1}, 60%, 55%), hsl(${h2}, 70%, 45%))`;
}

function getInitial(displayName: string): string {
  return displayName.charAt(0).toUpperCase();
}

export default function Profile() {
  const { username } = useParams<{ username: string }>();
  const { user: currentUser, setUser, logout } = useAuth();
  const navigate = useNavigate();
  const { data: profile, isLoading, error } = useUserProfile(username!);
  const updateProfile = useUpdateProfile();
  const changePassword = useChangePassword();

  const isOwnProfile = currentUser?.username === username;
  const [activeTab, setActiveTab] = useState<Tab>('posts');

  const [editingProfile, setEditingProfile] = useState(false);
  const [editingPassword, setEditingPassword] = useState(false);

  const [displayName, setDisplayName] = useState('');
  const [editUsername, setEditUsername] = useState('');
  const [bio, setBio] = useState('');
  const [profileError, setProfileError] = useState('');
  const [profileSuccess, setProfileSuccess] = useState('');

  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [passwordError, setPasswordError] = useState('');
  const [passwordSuccess, setPasswordSuccess] = useState('');

  // Tab data — only fetch when tab is active or is own profile
  const { data: postsData } = useUserPosts(username!);
  const { data: publicComments } = useUserComments(username!);
  const { data: likedPosts } = useMyLikedPosts(isOwnProfile && activeTab === 'liked');
  const { data: bookmarkedPosts } = useMyBookmarkedPosts(isOwnProfile && activeTab === 'bookmarks');
  const { data: myComments } = useMyComments(isOwnProfile && activeTab === 'comments');
  const { data: drafts } = useDrafts();

  const openEditProfile = () => {
    setDisplayName(profile?.display_name || '');
    setEditUsername(profile?.username || '');
    setBio(profile?.bio || '');
    setProfileError('');
    setProfileSuccess('');
    setEditingProfile(true);
    setEditingPassword(false);
  };

  const openEditPassword = () => {
    setCurrentPassword('');
    setNewPassword('');
    setConfirmPassword('');
    setPasswordError('');
    setPasswordSuccess('');
    setEditingPassword(true);
    setEditingProfile(false);
  };

  const handleProfileSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setProfileError('');
    setProfileSuccess('');

    if (!displayName.trim() && !editUsername.trim()) {
      setProfileError('Display name or username is required');
      return;
    }

    try {
      const updated = await updateProfile.mutateAsync({
        display_name: displayName.trim() || undefined,
        username: editUsername.trim() || undefined,
        bio: bio.trim() || undefined,
      });
      if (setUser) setUser(updated);
      setProfileSuccess('Profile updated successfully');
      setEditingProfile(false);
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        setProfileError('That username is already taken');
      } else {
        setProfileError(err instanceof Error ? err.message : 'Failed to update profile');
      }
    }
  };

  const handlePasswordSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setPasswordError('');
    setPasswordSuccess('');

    if (newPassword.length < 8) {
      setPasswordError('New password must be at least 8 characters');
      return;
    }
    if (newPassword !== confirmPassword) {
      setPasswordError('Passwords do not match');
      return;
    }

    try {
      await changePassword.mutateAsync({ current_password: currentPassword, new_password: newPassword });
      setPasswordSuccess('Password changed successfully');
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
      setEditingPassword(false);
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        setPasswordError('Current password is incorrect');
      } else {
        setPasswordError(err instanceof Error ? err.message : 'Failed to change password');
      }
    }
  };

  if (isLoading) return <div className={styles.loading}>Loading...</div>;
  if (error || !profile) return (
    <div className={styles.notFoundContainer}>
      <h2>User not found</h2>
      <p>The profile you're looking for doesn't exist.</p>
      <Link to="/" className={styles.backLink}>Back to home</Link>
    </div>
  );

  const posts = postsData?.data || [];
  const comments = isOwnProfile ? (myComments || []) : (publicComments || []);
  const tabs: { id: Tab; label: string; count?: number }[] = [
    { id: 'posts', label: 'Posts', count: posts.length },
    { id: 'comments', label: 'Comments', count: comments.length },
    ...(isOwnProfile ? [
      { id: 'liked' as Tab, label: 'Liked', count: likedPosts?.length },
      { id: 'bookmarks' as Tab, label: 'Bookmarks', count: bookmarkedPosts?.length },
      { id: 'drafts' as Tab, label: 'Drafts', count: drafts?.length },
    ] : []),
  ];

  return (
    <div className={styles.page}>
      {/* Banner */}
      <div
        className={styles.banner}
        style={profile.banner_url
          ? { backgroundImage: `url(${profile.banner_url})` }
          : { background: getBannerGradient(profile.username) }
        }
      />

      <div className={styles.container}>
        {/* Avatar + name row */}
        <div className={styles.profileHeader}>
          <div className={styles.avatarWrapper}>
            {profile.avatar_url ? (
              <img src={profile.avatar_url} alt={profile.display_name} className={styles.avatar} />
            ) : (
              <div className={styles.avatarFallback} style={{ background: getBannerGradient(profile.username) }}>
                {getInitial(profile.display_name)}
              </div>
            )}
          </div>

          <div className={styles.nameSection}>
            <h1 className={styles.displayName}>{profile.display_name}</h1>
            <p className={styles.usernameTag}>@{profile.username}</p>
            {profile.bio && <p className={styles.bio}>{profile.bio}</p>}
          </div>

          {isOwnProfile && (
            <div className={styles.editActions}>
              <button className={styles.editProfileBtn} onClick={openEditProfile}>
                Edit Profile
              </button>
              <button className={styles.changePasswordBtn} onClick={openEditPassword}>
                Change Password
              </button>
              <button
                className={styles.logoutBtn}
                onClick={async () => {
                  await logout();
                  navigate('/');
                }}
              >
                Logout
              </button>
            </div>
          )}
        </div>

        {/* Success messages outside forms */}
        {profileSuccess && <div className={styles.successBanner}>{profileSuccess}</div>}
        {passwordSuccess && <div className={styles.successBanner}>{passwordSuccess}</div>}

        {/* Edit Profile Form */}
        {editingProfile && (
          <div className={styles.editSection}>
            <h2 className={styles.editTitle}>Edit Profile</h2>
            <form onSubmit={handleProfileSave} className={styles.editForm}>
              {profileError && <div className={styles.formError}>{profileError}</div>}

              <div className={styles.field}>
                <label className={styles.label}>Display Name</label>
                <input
                  className={styles.input}
                  type="text"
                  value={displayName}
                  onChange={e => setDisplayName(e.target.value)}
                  placeholder="Your display name"
                />
              </div>

              <div className={styles.field}>
                <label className={styles.label}>Username</label>
                <input
                  className={styles.input}
                  type="text"
                  value={editUsername}
                  onChange={e => setEditUsername(e.target.value)}
                  placeholder="username"
                />
              </div>

              <div className={styles.field}>
                <label className={styles.label}>Bio</label>
                <textarea
                  className={styles.textarea}
                  value={bio}
                  onChange={e => setBio(e.target.value)}
                  placeholder="Tell us about yourself..."
                  rows={3}
                />
              </div>

              <div className={styles.formActions}>
                <button type="button" className={styles.cancelBtn} onClick={() => setEditingProfile(false)}>
                  Cancel
                </button>
                <button type="submit" className={styles.saveBtn} disabled={updateProfile.isPending}>
                  {updateProfile.isPending ? 'Saving...' : 'Save Changes'}
                </button>
              </div>
            </form>
          </div>
        )}

        {/* Change Password Form */}
        {editingPassword && (
          <div className={styles.editSection}>
            <h2 className={styles.editTitle}>Change Password</h2>
            <form onSubmit={handlePasswordSave} className={styles.editForm}>
              {passwordError && <div className={styles.formError}>{passwordError}</div>}

              <div className={styles.field}>
                <label className={styles.label}>Current Password</label>
                <input
                  className={styles.input}
                  type="password"
                  value={currentPassword}
                  onChange={e => setCurrentPassword(e.target.value)}
                  placeholder="Current password"
                  autoComplete="current-password"
                />
              </div>

              <div className={styles.field}>
                <label className={styles.label}>New Password</label>
                <input
                  className={styles.input}
                  type="password"
                  value={newPassword}
                  onChange={e => setNewPassword(e.target.value)}
                  placeholder="At least 8 characters"
                  autoComplete="new-password"
                />
              </div>

              <div className={styles.field}>
                <label className={styles.label}>Confirm New Password</label>
                <input
                  className={styles.input}
                  type="password"
                  value={confirmPassword}
                  onChange={e => setConfirmPassword(e.target.value)}
                  placeholder="Repeat new password"
                  autoComplete="new-password"
                />
              </div>

              <div className={styles.formActions}>
                <button type="button" className={styles.cancelBtn} onClick={() => setEditingPassword(false)}>
                  Cancel
                </button>
                <button type="submit" className={styles.saveBtn} disabled={changePassword.isPending}>
                  {changePassword.isPending ? 'Changing...' : 'Change Password'}
                </button>
              </div>
            </form>
          </div>
        )}

        {/* Tab nav */}
        <div className={styles.tabNav}>
          {tabs.map(tab => (
            <button
              key={tab.id}
              className={`${styles.tabBtn} ${activeTab === tab.id ? styles.tabActive : ''}`}
              onClick={() => setActiveTab(tab.id)}
            >
              {tab.label}
              {tab.count !== undefined && (
                <span className={styles.tabCount}>{tab.count}</span>
              )}
            </button>
          ))}
        </div>

        {/* Tab content */}
        <div className={styles.tabContent}>
          {activeTab === 'posts' && (
            posts.length === 0 ? (
              <p className={styles.empty}>No published posts yet.</p>
            ) : (
              <div className={styles.postGrid}>
                {posts.map(post => <PostCard key={post.id} post={post} />)}
              </div>
            )
          )}

          {activeTab === 'comments' && (
            comments.length === 0 ? (
              <p className={styles.empty}>No comments yet.</p>
            ) : (
              <div className={styles.commentList}>
                {comments.map(comment => (
                  <div key={comment.id} className={styles.commentCard}>
                    {comment.post_slug && (
                      <Link to={`/posts/${comment.post_slug}`} className={styles.commentPostLink}>
                        {comment.post_title || comment.post_slug}
                      </Link>
                    )}
                    <p className={styles.commentContent}>{comment.content}</p>
                    <span className={styles.commentDate}>
                      {new Date(comment.created_at).toLocaleDateString('en-US', {
                        month: 'short', day: 'numeric', year: 'numeric',
                      })}
                    </span>
                  </div>
                ))}
              </div>
            )
          )}

          {activeTab === 'liked' && isOwnProfile && (
            !likedPosts || likedPosts.length === 0 ? (
              <p className={styles.empty}>No liked posts yet.</p>
            ) : (
              <div className={styles.postGrid}>
                {likedPosts.map(post => <PostCard key={post.id} post={post} />)}
              </div>
            )
          )}

          {activeTab === 'bookmarks' && isOwnProfile && (
            !bookmarkedPosts || bookmarkedPosts.length === 0 ? (
              <p className={styles.empty}>No bookmarked posts yet.</p>
            ) : (
              <div className={styles.postGrid}>
                {bookmarkedPosts.map(post => <PostCard key={post.id} post={post} showBookmark />)}
              </div>
            )
          )}

          {activeTab === 'drafts' && isOwnProfile && (
            !drafts || drafts.length === 0 ? (
              <p className={styles.empty}>No drafts yet.</p>
            ) : (
              <div className={styles.postGrid}>
                {drafts.map(post => <PostCard key={post.id} post={post} />)}
              </div>
            )
          )}
        </div>
      </div>
    </div>
  );
}
