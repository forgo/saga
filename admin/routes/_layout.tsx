import { define } from "../utils.ts";
import { Layout } from "../components/Layout.tsx";

export default define.page(function AdminLayout(ctx) {
  return (
    <Layout currentPath={ctx.url.pathname}>
      <ctx.Component />
    </Layout>
  );
});
