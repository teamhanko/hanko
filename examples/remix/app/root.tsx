import type { LinksFunction, MetaFunction } from "@remix-run/node";
import { json } from "@remix-run/node";
import {
  Links,
  LiveReload,
  Meta,
  Outlet,
  Scripts,
  ScrollRestoration,
  useLoaderData,
} from "@remix-run/react";
import { createHead } from 'remix-island';
import styles from "~/styles/index.css";

export const links: LinksFunction = () => {
  return [{ rel: "stylesheet", href: styles }];
};

export const meta: MetaFunction = () => ({
  charset: "utf-8",
  title: "New Remix App",
  viewport: "width=device-width,initial-scale=1",
});

export const Head = createHead(() => (
  <>
    <Meta />
    <Links />
  </>
));

export async function loader() {
  return json({
    ENV: {
      HANKO_URL: process.env.REMIX_APP_HANKO_API,
      NODE_ENV: process.env.NODE_ENV,
    },
  });
}

export default function App() {
  const { ENV } = useLoaderData<typeof loader>();

  return (
    <>
      <Head />
      <Outlet />
      <ScrollRestoration />
      <Scripts />
      {/* Add the URL of the Hanko API instance to the window object so that
        it can be accessed in `useEffect` etc */}
      <script
        dangerouslySetInnerHTML={{
          __html: `window.ENV = ${JSON.stringify(ENV)}`,
        }}
      />
      <LiveReload />
    </>
  );
}

// Add typings to the window object so that TypeScript knows about the ENV variable we
// added
declare global {
  interface Window {
    ENV: {
      HANKO_URL: string;
    };
  }
}
