const preventEmpty = (id) => {
  if (document.getElementById(id).value === '') {
    return false;
  }
  return true;
};
