import {AppProps} from "$fresh/server.ts";
import {Head} from "$fresh/runtime.ts";

export default function App({Component}: AppProps) {
  return (
    <>
      <Head>
        <title>Hanko App</title>
        <link rel="stylesheet" href="/app.css"/>
      </Head>
      <div class="p-4 mx-auto max-w-screen-md">
        <Component/>
      </div>
    </>
  );
}
