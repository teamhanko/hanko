import type {NextPage} from 'next'
import dynamic from 'next/dynamic'
import styles from "../styles/Todo.module.css"

const HankoAuth = dynamic(() => import('../components/HankoAuth'), {
  ssr: false,
})

const Home: NextPage = () => {
  return (
    <div className={styles.content}>
      <HankoAuth/>
    </div>
  )
}

export default Home
