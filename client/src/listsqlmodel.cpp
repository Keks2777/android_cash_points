#include "listsqlmodel.h"

#include <QtCore/QJsonDocument>
#include <QtCore/QJsonParseError>
#include <QtCore/QJsonObject>
#include <QtCore/QDebug>

#define DEFAULT_ATTEMPTS_COUNT 3
#define DEFAULT_BATCH_SIZE 256

ListSqlModel::ListSqlModel(const QString &connectionName,
                           ServerApi *api,
                           IcoImageProvider *imageProvider,
                           QSettings *settings)
    : mApi(api),
      mImageProvider(imageProvider),
      mSettings(settings),
      mDBConnectionName(connectionName),
      mExpectedUploadCount(0),
      mUploadedCount(0)
{
    Q_ASSERT_X(api, "ListSqlModel(const QString &, ServerApi *, IcoImageProvider *, QSettings *)", "null ServerApi ptr");
    Q_ASSERT_X(imageProvider, "ListSqlModel(const QString &, ServerApi *, IcoImageProvider *, QSettings *)", "null IcoImageProvider ptr");
    Q_ASSERT_X(settings, "ListSqlModel(const QString &, ServerApi *, IcoImageProvider *, QSettings *)", "null QSettiings ptr");

    setAttemptsCount(DEFAULT_ATTEMPTS_COUNT);
    setRequestBatchSize(DEFAULT_BATCH_SIZE);

    setRoleName(SelectedRole, "selected");

    connect(this, SIGNAL(filterRequest(QString, QString)),
            this, SLOT(_setFilter(QString, QString)), Qt::QueuedConnection);
}

ListSqlModel::ListSqlModel(ListSqlModel *submodel)
    : mDBConnectionName(submodel->getDBConnectionName())
{
    Q_ASSERT_X(submodel, "ListSqlModel(ListSqlModel *)", "null ListSqlModel ptr");

    mApi = submodel->getServerApi();
    mImageProvider = submodel->getIcoImageProvider();
    mSettings = submodel->getSettings();

    setAttemptsCount(DEFAULT_ATTEMPTS_COUNT);
    setRequestBatchSize(DEFAULT_BATCH_SIZE);

    connect(this, SIGNAL(filterRequest(QString,QString)),
            this, SLOT(_setFilter(QString,QString)), Qt::QueuedConnection);
}

QString ListSqlModel::escapeFilter(QString filter)
{
    if (filter.isEmpty()) {
//        return "";
    }

    filter.replace('_', "");
    filter.replace('%', "");
    filter.replace('*', '%');
    filter.replace('?', '_');

    if (!filter.startsWith('%'))
    {
        filter.prepend('%');
    }

    if (!filter.endsWith('%'))
    {
        filter.append('%');
    }

    return filter;
}

void ListSqlModel::setFilter(QString filter, QString options)
{
    if (needEscapeFilter()) {
        emit filterRequest(escapeFilter(filter), options);
    } else {
        emit filterRequest(filter, options);
    }
}

void ListSqlModel::_setFilter(QString filter, QString options)
{
    QJsonParseError err;
    auto doc = QJsonDocument::fromJson(options.toUtf8(), &err);
    QJsonObject obj;
    if (err.error != QJsonParseError::NoError) {
        qWarning() << "Cannot parse filter options json!";
    } else {
        if (doc.isObject()) {
            obj = doc.object();
        } else {
            qWarning() << "Filter options must be json object";
        }
    }
    setFilterImpl(filter, obj);
}

void ListSqlModel::updateFromServer()
{
    updateFromServerImpl(getAttemptsCount());
}

bool ListSqlModel::setData(const QModelIndex &index, const QVariant &value, int role)
{
    return QStandardItemModel::setData(index, value, role);
}

QVariant ListSqlModel::data(const QModelIndex &item, int role) const
{
    if (role == getLastRole()) {
        return item.row();
    }

    return QStandardItemModel::data(item, role);
}

QHash<int, QByteArray> ListSqlModel::roleNames() const
{
    int lastRole = getLastRole();
    if (!mRoleNames.contains(lastRole)) {
        setRoleName(lastRole, "index");
    }

    return mRoleNames;
}
