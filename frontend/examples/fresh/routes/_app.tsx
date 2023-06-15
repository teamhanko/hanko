import { AppProps } from "$fresh/server.ts";
import { Head } from "$fresh/runtime.ts";

export default function App({ Component }: AppProps) {
  return (
    <>
      <Head>
        <link rel="stylesheet" href="/base.css" />
      </Head>
      <div class="wrapper">
        <Component />
      </div>
    </>
  );
}
