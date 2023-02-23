import React, {type ReactNode} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import {
  findFirstCategoryLink,
  useDocById,
} from '@docusaurus/theme-common/internal';
import isInternalUrl from '@docusaurus/isInternalUrl';
import {translate} from '@docusaurus/Translate';
import useBaseUrl from '@docusaurus/useBaseUrl';
import type {Props} from '@theme/DocCard';

import styles from './styles.module.css';
import type {
  PropSidebarItemCategory,
  PropSidebarItemLink,
} from '@docusaurus/plugin-content-docs';

function CardContainer({
  href,
  children,
}: {
  href: string;
  children: ReactNode;
}): JSX.Element {
  return (
    <Link
      href={href}
      className={clsx('card padding--lg', styles.cardContainer)}>
      {children}
    </Link>
  );
}

function CardLayout({
  href,
  icon,
  title,
  description,
  showDescription
}: {
  href: string;
  icon: ReactNode;
  title: string;
  description?: string;
  showDescription?: boolean;
}): JSX.Element {
  return (
    <CardContainer href={href}>
      <h2 className={clsx('text--truncate', styles.cardTitle)} title={title}>
        {icon} {title}
      </h2>
      {description && showDescription && (
        <p
          className={clsx('text--truncate', styles.cardDescription)}
          title={description}>
          {description}
        </p>
      )}
    </CardContainer>
  );
}

function CardCategory({
  item,
}: {
  item: PropSidebarItemCategory;
}): JSX.Element | null {
  const href = findFirstCategoryLink(item);

  // Unexpected: categories that don't have a link have been filtered upfront
  if (!href) {
    return null;
  }

  const icon = <DocCardIcon item={item}/>

  return (
    <CardLayout
      href={href}
      icon={icon}
      title={item.label}
      description={translate(
        {
          message: '{count} items',
          id: 'theme.docs.DocCard.categoryDescription',
          description:
            'The default description for a category card in the generated index about how many items this category includes',
        },
        {count: item.items.length},
      )}
      showDescription={!!item.customProps?.docCardShowDescription}
    />
  );
}

function CardLink({item}: {item: PropSidebarItemLink}): JSX.Element {
  const icon = <DocCardIcon item={item}/>

  const doc = useDocById(item.docId ?? undefined);
  return (
    <CardLayout
      href={item.href}
      icon={icon}
      title={item.label}
      description={doc?.description}
      showDescription={!!item.customProps?.docCardShowDescription}
    />
  );
}

export default function DocCard({item}: Props): JSX.Element {
  switch (item.type) {
    case 'link':
      return <CardLink item={item} />;
    case 'category':
      return <CardCategory item={item} />;
    default:
      throw new Error(`unknown item type ${JSON.stringify(item)}`);
  }
}

function DocCardIcon({ item }: {item: PropSidebarItemLink | PropSidebarItemCategory}): JSX.Element {
  let icon: ReactNode;

  if (item.customProps?.docCardIconName) {
    const iconName = item.customProps.docCardIconName;
    icon = <img alt={`${iconName} icon`} src={useBaseUrl(`/img/icons/${iconName}.svg`)}/>
  } else {
    icon = item.type === 'category' ? '🗃️' : isInternalUrl(item.href) ? '📄️' : '🔗';
  }

  return (
    <div className={styles.cardIcon}>
      {icon}
    </div>
  )
}
