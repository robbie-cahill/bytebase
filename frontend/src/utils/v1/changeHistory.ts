import slug from "slug";
import { UNKNOWN_ID } from "@/types";
import { ChangeHistory } from "@/types/proto/v1/database_service";
import { extractDatabaseResourceName } from "./database";
import { useDBSchemaV1Store, useDatabaseV1Store } from "@/store";
import { isUndefined, orderBy } from "lodash-es";
import { AffectedTable } from "@/types/changeHistory";

export const extractChangeHistoryUID = (changeHistorySlug: string) => {
  const parts = changeHistorySlug.split("-");
  return parts[parts.length - 1] ?? String(UNKNOWN_ID);
};

export const extractDatabaseNameAndChangeHistoryUID = (
  changeHistoryName: string
) => {
  const parts = changeHistoryName.split("/changeHistories/");
  if (parts.length !== 2) {
    throw new Error("Invalid change history name");
  }
  return {
    databaseName: parts[0],
    uid: extractChangeHistoryUID(parts[1]),
  };
};

export const changeHistorySlug = (uid: string, version: string): string => {
  return [slug(version), uid].join("-");
};

export const changeHistoryLink = (changeHistory: ChangeHistory): string => {
  const { name, uid, version } = changeHistory;
  const { instance, database } = extractDatabaseResourceName(name);
  const parent = `instances/${instance}/databases/${database}`;
  return `/${parent}/changeHistories/${changeHistorySlug(uid, version)}`;
};

export const getAffectedTablesOfChangeHistory = (
  changeHistory: ChangeHistory
): AffectedTable[] => {
  const { databaseName } = extractDatabaseNameAndChangeHistoryUID(
    changeHistory.name
  );
  const database = useDatabaseV1Store().getDatabaseByName(databaseName);
  const metadata = useDBSchemaV1Store().getDatabaseMetadata(database.name);
  return orderBy(
    changeHistory.changedResources?.databases
      .find((db) => db.name === database.databaseName)
      ?.schemas.map((schema) => {
        return schema.tables.map((table) => {
          const dropped = isUndefined(
            metadata.schemas
              .find((s) => s.name === schema.name)
              ?.tables.find((t) => t.name === table.name)
          );
          return {
            schema: schema.name,
            table: table.name,
            dropped,
          };
        });
      })
      .flat() || [],
    ["dropped"]
  );
};
