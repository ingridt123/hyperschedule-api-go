package course

import (
  "fmt"
  "regexp"
)

const expectHead = "" +
`"externalId",` +
`"classificationOfInstructionalProgramCode",` +
`"courseNumber",` +
`"courseTitle",` +
`"subjectAbbreviation",` +
`"departmentExternalId",` +
`"institutionExternalId",` +
`"facilityExternalId",` +
`"description"` + "\n"

var reHead = regexp.MustCompile(`^.*\n`)
var reStart = regexp.MustCompile(`(?m)^"([^"]+)","`)
var reRow = regexp.MustCompile(fmt.Sprintf(
  `^"%s","%s","%s","%s","%s","%s","%s","%s","%s"\n?$`,
  `([^"]+)`,    // externalId
  `([^"]*)`,    // classificationOfInstructionalProgramCode
  `([^"]+)`,    // courseNumber (same as externalId)
  `(.*?)`,      // courseTitle
  `([A-Z]*)`,   // subjectAbbreviation
  `([A-Z]*)`,   // departmentExternalId (same as subjectAbbreviation)
  `([A-Z]{2})`, // institutionExternalId
  `([A-Z]{2})`, // facilityExternalId (same as institutionExternalId)
  `((?s).*?)`,  // description
))
var reCode = regexp.MustCompile(`^([A-Z_/-]+) *([0-9A-Z/ -]*?) *([A-Z]{2})?$`)

func parse(row string) (*Entry, []*warn, *fail) {
  warns := make([]*warn, 0, 1)

  match := reRow.FindStringSubmatch(row)
  if match == nil {
    return nil, nil, &fail{
      full: row,
      data: failFullMatch{},
    }
  }

  colExternalId := match[1]
  colCourseNumber := match[3]
  colCourseTitle := match[4]
  colSubjectAbbreviation := match[5]
  colDepartmentExternalId := match[6]
  colInstitutionExternalId := match[7]
  colFacilityExternalId := match[8]
  colDescription := match[9]

  failWith := func(data failData) (*Entry, []*warn, *fail) {
    return nil, nil, &fail{
      full: row,
      id:   colExternalId,
      data: data,
    }
  }
  warnWith := func(data warnData) {
    warns = append(warns, &warn{
      full: row,
      id: colExternalId,
      data: data,
    })
  }

  if colExternalId != colCourseNumber {
    return failWith(&failCodeConsistent{
      externalId:   colExternalId,
      courseNumber: colCourseNumber,
    })
  }

  if colSubjectAbbreviation != colDepartmentExternalId {
    return failWith(&failSubjectConsistent{
      subjectAbbreviation:  colSubjectAbbreviation,
      departmentExternalId: colDepartmentExternalId,
    })
  }

  if colInstitutionExternalId != colFacilityExternalId {
    return failWith(&failCampusConsistent{
      institutionExternalId: colInstitutionExternalId,
      facilityExternalId:    colFacilityExternalId,
    })
  }

  matchCode := reCode.FindStringSubmatch(colExternalId)
  if matchCode == nil {
    return failWith(&failCodeMatch{
      input: match[1],
    })
  }

  codeDept := matchCode[1]
  codeNum := matchCode[2]
  codeCampus := matchCode[3]

  if len(codeCampus) == 0 {
    warnWith(&warnCodeCampusEmpty{
      externalId: match[1],
    })
  }

  return &Entry{
    Id:           colExternalId,
    Department:   codeDept,
    Number:       codeNum,
    DepartmentId: colDepartmentExternalId,
    Campus:       colInstitutionExternalId,
    Title:        colCourseTitle,
    Description:  colDescription,
  }, warns, nil
}

func ParseAll(contents []byte) ([]*Entry, []*warn, []*fail, error) {
  matchHead := string(reHead.Find(contents))
  if matchHead != expectHead {
    return nil, nil, nil, ErrIncorrectHead(matchHead)
  }
  contents = contents[len(matchHead):]

  chunks := append(reStart.FindAllIndex(contents, -1), []int{len(contents)})
  courses := make([]*Entry, 0, len(chunks)-1)
  fails := make([]*fail, 0, 8)
  warns := make([]*warn, 0, 8)
  for i, chunk := range chunks[:len(chunks)-1] {
    course, warn, fail := parse(string(contents[chunk[0]:chunks[i+1][0]]))
    if fail != nil {
      fails = append(fails, fail)
      continue
    }
    warns = append(warns, warn...)
    courses = append(courses, course)
  }
  return courses, warns, fails, nil
}
