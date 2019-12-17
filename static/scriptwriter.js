class __ScriptWriter {
  // from: http://www.writingroom.com/viewwriting/wr_how_to/How-To-Format-A-Screenplay
  SCREENPLAY_COMPONENTS = [
    "heading",
    "action",
    "character",
    "dialogue",
    "parenthetical",
    "transition"
  ];

  constructor(elementId) {
    // get dom element
    this.element = document.getElementById(elementId);

    // setup key bindings
    this.element.onkeydown = event => {
      if (!event.shiftKey && event.key === "Tab") {
        event.preventDefault();
        this.nextStyle();
      }
      if (event.shiftKey && event.key == "Tab") {
        event.preventDefault();
        this.prevStyle();
      }
    };
  }

  next = screenplayComponent => {
    // get the screenplay component after the current
    const index = this.SCREENPLAY_COMPONENTS.indexOf(screenplayComponent);
    if (index === -1) {
      return this.SCREENPLAY_COMPONENTS[0];
    }
    return this.SCREENPLAY_COMPONENTS[
      (index + 1) % this.SCREENPLAY_COMPONENTS.length
    ];
  };

  // get the screenplay component before the current
  // previous("action") -> "header"
  previous = screenplayComponent => {
    const index = this.SCREENPLAY_COMPONENTS.indexOf(screenplayComponent);
    if (index === -1) {
      return this.SCREENPLAY_COMPONENTS[0];
    }
    return this.SCREENPLAY_COMPONENTS[
      (index + this.SCREENPLAY_COMPONENTS.length - 1) %
        this.SCREENPLAY_COMPONENTS.length
    ];
  };

  getActiveElement = () => {
    let focused = window.getSelection().focusNode;
    return focused.nodeName === "#text" ? focused.parentNode : focused;
  };

  nextStyle = () => {
    document.execCommand("formatBlock", false, "p");
    const activeElement = this.getActiveElement();
    const nextScreenplayComponent = this.next(activeElement.className);
    activeElement.classList = nextScreenplayComponent;
  };

  prevStyle = () => {
    document.execCommand("formatBlock", false, "p");
    const activeElement = this.getActiveElement();
    const nextScreenplayComponent = this.previous(activeElement.className);
    activeElement.classList = nextScreenplayComponent;
  };
}

const scriptwriter = {
  edit: elementId => new __ScriptWriter(elementId)
};
