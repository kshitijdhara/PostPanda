import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import styles from './MarkdownRenderer.module.scss';

interface Props {
  content: string;
}

export default function MarkdownRenderer({ content }: Props) {
  return (
    <div className={styles.markdown}>
      <ReactMarkdown remarkPlugins={[remarkGfm]}>{content}</ReactMarkdown>
    </div>
  );
}
