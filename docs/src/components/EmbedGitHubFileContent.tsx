import React, { FC } from "react";
import CodeBlock from '@theme/CodeBlock';

function convertGitHubUrlToRaw(url: string): string {
  const [user, repo, blob, branch, ...filePath] = url.replace("https://github.com/", "").split("/");
  const rawFileUrl = `https://raw.githubusercontent.com/${user}/${repo}/${branch}/${filePath.join("/")}`;
  return rawFileUrl;
}

function getFileExtension(url: string): string {
  const filename = url.split("/").pop();
  return filename.slice((Math.max(0, filename.lastIndexOf(".")) || Infinity) + 1);
}

function getFilePath(url: string): string {
  return url.replace("https://github.com/", "").split("/").slice(4, Infinity).join("/");
}

type EmbedGitHubFileContentProps = {
  url: string,
  loadingComponent?: JSX.Element,
  errorComponent?: JSX.Element,
  onLoad?: () => void
  onError?: (e: Error) => void
}

const EmbedGitHubFileContent: FC<EmbedGitHubFileContentProps> = ({
  url,
  loadingComponent,
  errorComponent,
  onLoad,
  onError
}) => {
  const [gitHubFileContent, setGitHubFileContent] = React.useState("");
  const [isLoading, setIsLoading] = React.useState(true);
  const [errorOccurred, setErrorOccurred] = React.useState(false);

  const gitHubUrlRegEx = /^https:\/\/github\.com\/.*\/.*\/blob\/.*/;
  if (!url.match(gitHubUrlRegEx)) {
    throw new Error("Invalid URL format");
  }

  React.useEffect(() => {
    setGitHubFileContent("");
    setIsLoading(true);
    setErrorOccurred(false);

    const rawFileUrl = convertGitHubUrlToRaw(url);

    const fetchGitHubFileContent = async () => {
      try {
        const response = await fetch(rawFileUrl, {
          headers: {
            "Content-Type": "text/plain; charset=utf-8",
          },
        });

        if (!response.ok) {
          setErrorOccurred(true);
          onError(new Error());
        } else {
          const text = await response.text();
          setGitHubFileContent(text);
          setIsLoading(false);
          onLoad();
        }
      } catch (err) {
        setErrorOccurred(true);
        onError(err);
      }
    };

    fetchGitHubFileContent();
  }, [url, onLoad, onError]);

  if (errorOccurred) {
    return errorComponent;
  }

  if (isLoading) {
    return loadingComponent;
  }

  return (
    <CodeBlock language={getFileExtension(url)} title={getFilePath(url)}>{gitHubFileContent}</CodeBlock>
  );
}

EmbedGitHubFileContent.defaultProps = {
  loadingComponent: <p>loading...</p>,
  errorComponent: <p>an error occured.</p>,
  onLoad: () => {},
  onError: (e) => console.log(e)
};

export default EmbedGitHubFileContent;
