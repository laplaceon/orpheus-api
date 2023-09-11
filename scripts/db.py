import pymysql
from threading import Lock
from sqlalchemy.engine.url import make_url

date_format = "%Y-%m-%dT%H:%M:%S%z"

class Database():
    def __init__(self, url):

        url = make_url(f"mysql://{url}")

        connection = pymysql.connect(host=url.host[4:-1],
                                database=url.database,
                                user=url.username,
                                password=url.password,
                                ssl={"rejectUnauthorized": True})

        if connection is not None:
            connection.autocommit(True)
            self.connection = connection
            self.cursor = self.connection.cursor()
            self.mutex = Lock()
        else:
            raise Exception("Connection is not valid.")

    def save_urls_to_history(self, historyid, urls):
        # for url in urls:
        #     self.cursor.execute()
        return 1


    def get_history_item(self, historyId):
        self.mutex.acquire()
        self.cursor.execute("""SELECT expiry_hours, p.status FROM plans 
            JOIN (SELECT plan_id, status FROM history WHERE id = %s) p 
            ON plans.id = p.plan_id;""", historyId)

        result = self.cursor.fetchone()

        self.mutex.release()

        return result

    def get_assets(self):
        self.cursor.execute('select id, symbol, class from assets where enabled = 1')
        results = self.cursor.fetchall()

        self.cursor.execute('select asset_id, alias, queryable, matchable from asset_aliases')
        name_results = {}

        for x in self.cursor.fetchall():
            if x[0] in name_results:
                name_results[x[0]].append([x[1], x[2], x[3]])
            else:
                name_results[x[0]] = [[x[1], x[2], x[3]]]

        assets = []

        for r in results:
            match_terms = []
            query_terms = []
            if r[0] in name_results:
                match_terms = [x[0] for x in name_results[r[0]] if x[2] == 1]
                query_terms = [x[0] for x in name_results[r[0]] if x[1] == 1]
            assets.append({'id': r[0], 'symbol': r[1], 'class': r[2], 'match_terms': match_terms, 'query_terms': query_terms})

        return assets

    def get_sites(self):
        self.cursor.execute('select base_url, name, weight from sites')

        results = self.cursor.fetchall()

        return results

    def sentiment_is_unprocessed(self, sentiment_id):
        self.mutex.acquire()
        self.cursor.execute('SELECT EXISTS (SELECT * FROM sentiments WHERE id = %s AND processed_at IS NULL AND attempts_tried < 3)', sentiment_id)

        result = self.cursor.fetchone()

        self.mutex.release()

        # print('result', result, result[0])

        return result[0]

    def get_unprocessed_news_articles(self):
        self.cursor.execute('select s.id, s.asset_id, na.url from sentiments s join news_articles na on na.id = s.entity_id where entity_type = 0 and processed_at IS NULL and attempts_tried < 3')
        results = self.cursor.fetchall()
        return results

    def increment_attempts(self, id):
        self.mutex.acquire()
        result = self.cursor.execute('update sentiments set attempts_tried = attempts_tried + 1 where id = {}'.format(id))
        self.mutex.release()

        return result

    def update_sentiment_scores(self, id, sscore, rscore):
        self.mutex.acquire()
        result = self.cursor.execute('update sentiments set score = %s, relevance = %s, processed_at = CURRENT_TIMESTAMP() where id = %s',
            (float(sscore), float(rscore), id))
        self.mutex.release()

        return result

    def save_news_article(self, asset_id, url, pub_date):
        self.cursor.execute("insert ignore into news_articles (url, published_at) values (%s, %s)",
            (url, pub_date))

        self.cursor.execute("select id from news_articles where url = %s", url)
        article_id = self.cursor.fetchone()[0]

        self.cursor.execute("insert ignore into sentiments (asset_id, entity_id, entity_type) values (%s, %s, 0)",
            (asset_id, article_id))

        return True

    def save_log(self, type, status):
        result = self.cursor.execute("insert into logs (type, status) values (%s, %s)",
            (type, status))

        return True

    def quit(self):
        self.cursor.close()
        self.connection.close()