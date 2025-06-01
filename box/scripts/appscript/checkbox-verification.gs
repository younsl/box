function onEdit(e) {
  var sheet = e.source.getActiveSheet();
  var range = e.range;
  var value = range.getValue();
  
  // 수정이 발생한 셀의 행과 열 인덱스 가져오기
  var row = range.getRow();
  var column = range.getColumn();
  
  // 체크 상태에 따른 색상 변수 정의
  var checkedColor = "#98FB98"; // 체크된 경우의 색상 (연두색)
  var uncheckedColor = "#FFC0CB"; // 체크되지 않은 경우의 색상 (핑크색)
  
  // B열에서 수정이 발생한 경우
  if (column == 2) {
    // 해당 행의 A부터 C 컬럼 범위 설정
    var rangeToFormat = sheet.getRange(row, 1, 1, 3); // A부터 C 컬럼의 범위
    
    // 체크박스가 체크된 경우
    if (value === true) {
      rangeToFormat.setBackground(checkedColor); // 체크된 상태의 색상으로 설정
    } else {
      rangeToFormat.setBackground(uncheckedColor); // 체크되지 않은 상태의 색상으로 설정
    }
  }
}
