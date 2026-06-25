import { Upload } from "lucide-react";

export function EmptyState({ onImport }: { onImport: () => void }) {
  return (
    <div className="empty-state">
      <Upload size={42} />
      <h2>还没有比赛数据</h2>
      <p>导入包含本帮会与对手帮会的联赛 CSV 后，这里会生成排名、对比和六维分析。</p>
      <button className="btn primary" onClick={onImport}>
        导入 CSV
      </button>
    </div>
  );
}
