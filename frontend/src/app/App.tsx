import { Routes, Route } from 'react-router-dom';
import Layout from '../components/Layout/Layout';
import ProtectedRoute from '../components/ProtectedRoute/ProtectedRoute';
import PostList from '../pages/PostList';
import PostDetail from '../pages/PostDetail';
import PostCreate from '../pages/PostCreate';
import PostEdit from '../pages/PostEdit';
import DraftList from '../pages/DraftList';
import Login from '../pages/Login';
import Register from '../pages/Register';
import Profile from '../pages/Profile';

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<PostList />} />
        <Route path="/posts/:slug" element={<PostDetail />} />
        <Route
          path="/posts/new"
          element={
            <ProtectedRoute>
              <PostCreate />
            </ProtectedRoute>
          }
        />
        <Route
          path="/posts/:slug/edit"
          element={
            <ProtectedRoute>
              <PostEdit />
            </ProtectedRoute>
          }
        />
        <Route
          path="/drafts"
          element={
            <ProtectedRoute>
              <DraftList />
            </ProtectedRoute>
          }
        />
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/profile/:username" element={<Profile />} />
      </Route>
    </Routes>
  );
}
