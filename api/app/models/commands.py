"""
Pydantic models for all Virtuoso CLI commands.

This module contains comprehensive models for all 70+ CLI commands organized by command groups.
"""
from typing import Optional, List, Dict, Any, Union, Literal
from enum import Enum
from pydantic import BaseModel, Field, ConfigDict, field_validator


# ========================================
# Enums for constrained values
# ========================================


class StepType(str, Enum):
    """All supported step types in Virtuoso"""

    # Navigation
    NAVIGATE = "NAVIGATE"

    # Interaction
    CLICK = "CLICK"
    DOUBLE_CLICK = "DOUBLE_CLICK"
    RIGHT_CLICK = "RIGHT_CLICK"
    HOVER = "HOVER"
    FILL = "FILL"  # Write/Type
    KEY = "KEY"

    # Mouse
    MOUSE_MOVE = "MOUSE_MOVE"
    MOUSE_DOWN = "MOUSE_DOWN"
    MOUSE_UP = "MOUSE_UP"

    # Scroll
    SCROLL = "SCROLL"
    SCROLL_TOP = "SCROLL_TOP"
    SCROLL_BOTTOM = "SCROLL_BOTTOM"
    SCROLL_ELEMENT = "SCROLL_ELEMENT"
    SCROLL_POSITION = "SCROLL_POSITION"

    # Window
    WINDOW_RESIZE = "WINDOW_RESIZE"
    SWITCH_TAB = "SWITCH_TAB"
    SWITCH_FRAME = "SWITCH_FRAME"
    SWITCH_PARENT_FRAME = "SWITCH_PARENT_FRAME"

    # Wait
    WAIT = "WAIT"
    WAIT_FOR_ELEMENT = "WAIT_FOR_ELEMENT"

    # Assertions
    ASSERT_EXISTS = "ASSERT_EXISTS"
    ASSERT_NOT_EXISTS = "ASSERT_NOT_EXISTS"
    ASSERT_TEXT = "ASSERT_TEXT"  # equals
    ASSERT_NOT_TEXT = "ASSERT_NOT_TEXT"  # not-equals
    ASSERT_CONTAINS = "ASSERT_CONTAINS"
    ASSERT_NOT_CONTAINS = "ASSERT_NOT_CONTAINS"
    ASSERT_CHECKED = "ASSERT_CHECKED"
    ASSERT_NOT_CHECKED = "ASSERT_NOT_CHECKED"
    ASSERT_SELECTED = "ASSERT_SELECTED"
    ASSERT_NOT_SELECTED = "ASSERT_NOT_SELECTED"
    ASSERT_VARIABLE = "ASSERT_VARIABLE"
    ASSERT_GREATER_THAN = "ASSERT_GREATER_THAN"
    ASSERT_LESS_THAN = "ASSERT_LESS_THAN"
    ASSERT_MATCHES = "ASSERT_MATCHES"

    # Data
    STORE = "STORE"
    STORE_TEXT = "STORE_TEXT"

    # Cookies
    COOKIE_CREATE = "COOKIE_CREATE"
    COOKIE_DELETE = "COOKIE_DELETE"
    COOKIE_WIPE_ALL = "COOKIE_WIPE_ALL"

    # Dialog
    ALERT = "ALERT"
    CONFIRM = "CONFIRM"
    PROMPT = "PROMPT"

    # File
    UPLOAD = "UPLOAD"

    # Misc
    COMMENT = "COMMENT"
    EXECUTE_SCRIPT = "EXECUTE_SCRIPT"


class ElementType(str, Enum):
    """Common element types for targeting"""

    BUTTON = "BUTTON"
    LINK = "LINK"
    INPUT = "INPUT"
    CHECKBOX = "CHECKBOX"
    RADIO = "RADIO"
    SELECT = "SELECT"
    TEXTAREA = "TEXTAREA"
    DIV = "DIV"
    SPAN = "SPAN"
    IMAGE = "IMAGE"
    TABLE = "TABLE"
    FORM = "FORM"


class Position(str, Enum):
    """Element position options"""

    CENTER = "CENTER"
    TOP_LEFT = "TOP_LEFT"
    TOP_RIGHT = "TOP_RIGHT"
    BOTTOM_LEFT = "BOTTOM_LEFT"
    BOTTOM_RIGHT = "BOTTOM_RIGHT"


class MouseButton(str, Enum):
    """Mouse button options"""

    LEFT = "left"
    RIGHT = "right"
    MIDDLE = "middle"


class KeyModifier(str, Enum):
    """Keyboard modifier keys"""

    CTRL = "ctrl"
    SHIFT = "shift"
    ALT = "alt"
    META = "meta"


class ScrollDirection(str, Enum):
    """Scroll direction options"""

    UP = "up"
    DOWN = "down"
    LEFT = "left"
    RIGHT = "right"


class TabDirection(str, Enum):
    """Tab switch direction"""

    NEXT = "next"
    PREVIOUS = "previous"
    FIRST = "first"
    LAST = "last"


class OutputFormat(str, Enum):
    """Output format options"""

    JSON = "json"
    YAML = "yaml"
    HUMAN = "human"
    AI = "ai"


# ========================================
# Base Models
# ========================================


class BaseCommand(BaseModel):
    """Base model for all commands"""

    model_config = ConfigDict(use_enum_values=True)

    checkpoint_id: Optional[str] = Field(
        None, description="Checkpoint ID (can be omitted if using session context)"
    )
    position: Optional[int] = Field(
        None, description="Step position in the test sequence", ge=1
    )
    output_format: Optional[OutputFormat] = Field(
        OutputFormat.JSON, description="Output format"
    )


class SelectorCommand(BaseCommand):
    """Base for commands that use selectors"""

    selector: str = Field(
        ..., description="CSS selector or text to find element", min_length=1
    )


# ========================================
# Assertion Commands (step-assert)
# ========================================


class AssertExistsCommand(SelectorCommand):
    """Assert element exists command"""

    command: Literal["exists"] = "exists"

    class Config:
        json_schema_extra = {
            "example": {
                "selector": "button#submit",
                "checkpoint_id": "12345",
                "position": 1,
            }
        }


class AssertNotExistsCommand(SelectorCommand):
    """Assert element does not exist command"""

    command: Literal["not-exists"] = "not-exists"


class AssertEqualsCommand(SelectorCommand):
    """Assert element text equals value"""

    command: Literal["equals"] = "equals"
    value: str = Field(..., description="Expected text value")

    class Config:
        json_schema_extra = {
            "example": {
                "selector": "h1",
                "value": "Welcome",
                "checkpoint_id": "12345",
                "position": 2,
            }
        }


class AssertNotEqualsCommand(SelectorCommand):
    """Assert element text does not equal value"""

    command: Literal["not-equals"] = "not-equals"
    value: str = Field(..., description="Value that should not match")


class AssertCheckedCommand(SelectorCommand):
    """Assert checkbox/radio is checked"""

    command: Literal["checked"] = "checked"


class AssertSelectedCommand(SelectorCommand):
    """Assert option is selected"""

    command: Literal["selected"] = "selected"


class AssertVariableCommand(BaseCommand):
    """Assert variable value"""

    command: Literal["variable"] = "variable"
    variable_name: str = Field(..., description="Variable name (without $ prefix)")
    expected_value: str = Field(..., description="Expected variable value")


class AssertGreaterThanCommand(BaseCommand):
    """Assert value is greater than threshold"""

    command: Literal["gt"] = "gt"
    variable_name: str = Field(
        ..., description="Variable name containing numeric value"
    )
    threshold: Union[int, float] = Field(..., description="Threshold value")


class AssertGreaterThanOrEqualCommand(BaseCommand):
    """Assert value is greater than or equal to threshold"""

    command: Literal["gte"] = "gte"
    variable_name: str = Field(
        ..., description="Variable name containing numeric value"
    )
    threshold: Union[int, float] = Field(..., description="Threshold value")


class AssertLessThanCommand(BaseCommand):
    """Assert value is less than threshold"""

    command: Literal["lt"] = "lt"
    variable_name: str = Field(
        ..., description="Variable name containing numeric value"
    )
    threshold: Union[int, float] = Field(..., description="Threshold value")


class AssertLessThanOrEqualCommand(BaseCommand):
    """Assert value is less than or equal to threshold"""

    command: Literal["lte"] = "lte"
    variable_name: str = Field(
        ..., description="Variable name containing numeric value"
    )
    threshold: Union[int, float] = Field(..., description="Threshold value")


class AssertMatchesCommand(BaseCommand):
    """Assert value matches regex pattern"""

    command: Literal["matches"] = "matches"
    variable_name: str = Field(..., description="Variable name to test")
    pattern: str = Field(..., description="Regular expression pattern")


# Union type for all assertion commands
AssertCommand = Union[
    AssertExistsCommand,
    AssertNotExistsCommand,
    AssertEqualsCommand,
    AssertNotEqualsCommand,
    AssertCheckedCommand,
    AssertSelectedCommand,
    AssertVariableCommand,
    AssertGreaterThanCommand,
    AssertGreaterThanOrEqualCommand,
    AssertLessThanCommand,
    AssertLessThanOrEqualCommand,
    AssertMatchesCommand,
]


# ========================================
# Interaction Commands (step-interact)
# ========================================


class ClickCommand(SelectorCommand):
    """Click element command"""

    command: Literal["click"] = "click"
    modifier: Optional[KeyModifier] = Field(
        None, description="Keyboard modifier to hold during click"
    )

    class Config:
        json_schema_extra = {
            "example": {
                "selector": "button[type='submit']",
                "checkpoint_id": "12345",
                "position": 1,
            }
        }


class DoubleClickCommand(SelectorCommand):
    """Double-click element command"""

    command: Literal["double-click"] = "double-click"


class RightClickCommand(SelectorCommand):
    """Right-click element command"""

    command: Literal["right-click"] = "right-click"


class HoverCommand(SelectorCommand):
    """Hover over element command"""

    command: Literal["hover"] = "hover"


class WriteCommand(SelectorCommand):
    """Write text into element command"""

    command: Literal["write"] = "write"
    text: str = Field(..., description="Text to write")
    clear_first: Optional[bool] = Field(True, description="Clear existing text first")

    class Config:
        json_schema_extra = {
            "example": {
                "selector": "input#email",
                "text": "test@example.com",
                "checkpoint_id": "12345",
                "position": 2,
            }
        }


class KeyCommand(BaseCommand):
    """Send keyboard key command"""

    command: Literal["key"] = "key"
    key: str = Field(..., description="Key to press (e.g., Enter, Tab, Escape)")
    modifiers: Optional[List[KeyModifier]] = Field(
        None, description="Modifier keys to hold"
    )
    selector: Optional[str] = Field(None, description="Optional element to focus first")


# Mouse sub-commands
class MouseMoveToCommand(SelectorCommand):
    """Move mouse to element"""

    command: Literal["move-to"] = "move-to"


class MouseMoveByCommand(BaseCommand):
    """Move mouse by offset"""

    command: Literal["move-by"] = "move-by"
    x: int = Field(..., description="X offset in pixels")
    y: int = Field(..., description="Y offset in pixels")


class MouseDownCommand(BaseCommand):
    """Mouse button down"""

    command: Literal["down"] = "down"
    button: Optional[MouseButton] = Field(MouseButton.LEFT, description="Mouse button")


class MouseUpCommand(BaseCommand):
    """Mouse button up"""

    command: Literal["up"] = "up"
    button: Optional[MouseButton] = Field(MouseButton.LEFT, description="Mouse button")


# Select sub-commands
class SelectOptionCommand(SelectorCommand):
    """Select dropdown option by text"""

    command: Literal["option"] = "option"
    value: str = Field(..., description="Option text to select")


class SelectIndexCommand(SelectorCommand):
    """Select dropdown option by index"""

    command: Literal["index"] = "index"
    index: int = Field(..., description="Zero-based index", ge=0)


class SelectLastCommand(SelectorCommand):
    """Select last dropdown option"""

    command: Literal["last"] = "last"


# Union types for sub-commands
MouseCommand = Union[
    MouseMoveToCommand, MouseMoveByCommand, MouseDownCommand, MouseUpCommand
]
SelectCommand = Union[SelectOptionCommand, SelectIndexCommand, SelectLastCommand]

# Union type for all interaction commands
InteractCommand = Union[
    ClickCommand,
    DoubleClickCommand,
    RightClickCommand,
    HoverCommand,
    WriteCommand,
    KeyCommand,
    MouseCommand,
    SelectCommand,
]


# ========================================
# Navigation Commands (step-navigate)
# ========================================


class NavigateToCommand(BaseCommand):
    """Navigate to URL command"""

    command: Literal["to"] = "to"
    url: str = Field(..., description="URL to navigate to", pattern=r"^https?://")

    class Config:
        json_schema_extra = {
            "example": {
                "url": "https://example.com",
                "checkpoint_id": "12345",
                "position": 1,
            }
        }


class ScrollTopCommand(BaseCommand):
    """Scroll to top of page"""

    command: Literal["scroll-top"] = "scroll-top"


class ScrollBottomCommand(BaseCommand):
    """Scroll to bottom of page"""

    command: Literal["scroll-bottom"] = "scroll-bottom"


class ScrollElementCommand(SelectorCommand):
    """Scroll element into view"""

    command: Literal["scroll-element"] = "scroll-element"


class ScrollPositionCommand(BaseCommand):
    """Scroll to specific position"""

    command: Literal["scroll-position"] = "scroll-position"
    x: int = Field(..., description="X coordinate", ge=0)
    y: int = Field(..., description="Y coordinate", ge=0)


class ScrollByCommand(BaseCommand):
    """Scroll by offset"""

    command: Literal["scroll-by"] = "scroll-by"
    x: int = Field(0, description="X offset")
    y: int = Field(..., description="Y offset")


class ScrollUpCommand(BaseCommand):
    """Scroll up by pixels"""

    command: Literal["scroll-up"] = "scroll-up"
    pixels: int = Field(100, description="Pixels to scroll", gt=0)


class ScrollDownCommand(BaseCommand):
    """Scroll down by pixels"""

    command: Literal["scroll-down"] = "scroll-down"
    pixels: int = Field(100, description="Pixels to scroll", gt=0)


# Union type for all navigation commands
NavigateCommand = Union[
    NavigateToCommand,
    ScrollTopCommand,
    ScrollBottomCommand,
    ScrollElementCommand,
    ScrollPositionCommand,
    ScrollByCommand,
    ScrollUpCommand,
    ScrollDownCommand,
]


# ========================================
# Window Commands (step-window)
# ========================================


class WindowResizeCommand(BaseCommand):
    """Resize window command"""

    command: Literal["resize"] = "resize"
    dimensions: str = Field(
        ..., description="Window dimensions (e.g., '1024x768')", pattern=r"^\d+x\d+$"
    )

    @field_validator("dimensions")
    def validate_dimensions(cls, v):
        parts = v.split("x")
        width, height = int(parts[0]), int(parts[1])
        if width < 100 or height < 100:
            raise ValueError("Window dimensions must be at least 100x100")
        return v


class WindowMaximizeCommand(BaseCommand):
    """Maximize window command"""

    command: Literal["maximize"] = "maximize"


class SwitchTabCommand(BaseCommand):
    """Switch browser tab"""

    command: Literal["switch-tab"] = "switch-tab"
    direction: TabDirection = Field(..., description="Tab direction")


class SwitchIframeCommand(SelectorCommand):
    """Switch to iframe"""

    command: Literal["switch-iframe"] = "switch-iframe"


class SwitchParentFrameCommand(BaseCommand):
    """Switch to parent frame"""

    command: Literal["switch-parent-frame"] = "switch-parent-frame"


# Union type for all window commands
WindowCommand = Union[
    WindowResizeCommand,
    WindowMaximizeCommand,
    SwitchTabCommand,
    SwitchIframeCommand,
    SwitchParentFrameCommand,
]


# ========================================
# Data Commands (step-data)
# ========================================


class StoreTextCommand(SelectorCommand):
    """Store element text in variable"""

    command: Literal["store-text"] = "store-text"
    variable_name: str = Field(
        ..., description="Variable name (without $ prefix)", pattern=r"^[a-zA-Z_]\w*$"
    )


class StoreValueCommand(SelectorCommand):
    """Store element value in variable"""

    command: Literal["store-value"] = "store-value"
    variable_name: str = Field(
        ..., description="Variable name (without $ prefix)", pattern=r"^[a-zA-Z_]\w*$"
    )


class StoreAttributeCommand(SelectorCommand):
    """Store element attribute in variable"""

    command: Literal["store-attribute"] = "store-attribute"
    attribute_name: str = Field(..., description="Attribute name to store")
    variable_name: str = Field(
        ..., description="Variable name (without $ prefix)", pattern=r"^[a-zA-Z_]\w*$"
    )


class CookieCreateCommand(BaseCommand):
    """Create cookie command"""

    command: Literal["cookie-create"] = "cookie-create"
    name: str = Field(..., description="Cookie name")
    value: str = Field(..., description="Cookie value")
    domain: Optional[str] = Field(None, description="Cookie domain")
    path: Optional[str] = Field("/", description="Cookie path")
    secure: Optional[bool] = Field(False, description="Secure cookie flag")
    http_only: Optional[bool] = Field(False, description="HTTP only flag")


class CookieDeleteCommand(BaseCommand):
    """Delete cookie command"""

    command: Literal["cookie-delete"] = "cookie-delete"
    name: str = Field(..., description="Cookie name to delete")


class CookieClearCommand(BaseCommand):
    """Clear all cookies command"""

    command: Literal["cookie-clear"] = "cookie-clear"


# Union type for all data commands
DataCommand = Union[
    StoreTextCommand,
    StoreValueCommand,
    StoreAttributeCommand,
    CookieCreateCommand,
    CookieDeleteCommand,
    CookieClearCommand,
]


# ========================================
# Dialog Commands (step-dialog)
# ========================================


class DismissAlertCommand(BaseCommand):
    """Dismiss alert dialog"""

    command: Literal["dismiss-alert"] = "dismiss-alert"


class DismissConfirmCommand(BaseCommand):
    """Dismiss confirm dialog"""

    command: Literal["dismiss-confirm"] = "dismiss-confirm"
    accept: bool = Field(
        True, description="Accept (true) or reject (false) the confirmation"
    )


class DismissPromptCommand(BaseCommand):
    """Dismiss prompt dialog"""

    command: Literal["dismiss-prompt"] = "dismiss-prompt"
    accept: bool = Field(True, description="Accept (true) or reject (false) the prompt")


class DismissPromptWithTextCommand(BaseCommand):
    """Dismiss prompt dialog with text"""

    command: Literal["dismiss-prompt-with-text"] = "dismiss-prompt-with-text"
    text: str = Field(..., description="Text to enter in prompt")


# Union type for all dialog commands
DialogCommand = Union[
    DismissAlertCommand,
    DismissConfirmCommand,
    DismissPromptCommand,
    DismissPromptWithTextCommand,
]


# ========================================
# Wait Commands (step-wait)
# ========================================


class WaitElementCommand(SelectorCommand):
    """Wait for element to be visible"""

    command: Literal["element"] = "element"
    timeout: Optional[int] = Field(30000, description="Timeout in milliseconds", gt=0)


class WaitTimeCommand(BaseCommand):
    """Wait for specified time"""

    command: Literal["time"] = "time"
    milliseconds: int = Field(
        ..., description="Time to wait in milliseconds", gt=0, le=300000
    )


# Union type for all wait commands
WaitCommand = Union[WaitElementCommand, WaitTimeCommand]


# ========================================
# File Commands (step-file)
# ========================================


class FileUploadCommand(SelectorCommand):
    """Upload file from URL"""

    command: Literal["upload"] = "upload"
    url: str = Field(..., description="URL of file to upload", pattern=r"^https?://")


class FileUploadUrlCommand(SelectorCommand):
    """Upload file from URL (alias)"""

    command: Literal["upload-url"] = "upload-url"
    url: str = Field(..., description="URL of file to upload", pattern=r"^https?://")


# Union type for all file commands
FileCommand = Union[FileUploadCommand, FileUploadUrlCommand]


# ========================================
# Misc Commands (step-misc)
# ========================================


class CommentCommand(BaseCommand):
    """Add comment to test"""

    command: Literal["comment"] = "comment"
    text: str = Field(..., description="Comment text")


class ExecuteCommand(BaseCommand):
    """Execute JavaScript code"""

    command: Literal["execute"] = "execute"
    script: str = Field(..., description="JavaScript code to execute")

    class Config:
        json_schema_extra = {
            "example": {
                "script": "console.log('Test message');",
                "checkpoint_id": "12345",
                "position": 1,
            }
        }


# Union type for all misc commands
MiscCommand = Union[CommentCommand, ExecuteCommand]


# ========================================
# Library Commands
# ========================================


class LibraryAddCommand(BaseCommand):
    """Add step to library"""

    command: Literal["add"] = "add"
    checkpoint_id: str = Field(..., description="Checkpoint ID containing steps")
    position: int = Field(..., description="Position of step to add", ge=1)
    name: str = Field(..., description="Library step name")
    category: Optional[str] = Field(None, description="Library category")


class LibraryGetCommand(BaseCommand):
    """Get library step details"""

    command: Literal["get"] = "get"
    step_id: str = Field(..., description="Library step ID")


class LibraryAttachCommand(BaseCommand):
    """Attach library step to checkpoint"""

    command: Literal["attach"] = "attach"
    library_step_id: str = Field(..., description="Library step ID")
    checkpoint_id: str = Field(..., description="Target checkpoint ID")
    position: Optional[int] = Field(None, description="Position in checkpoint", ge=1)


class LibraryMoveStepCommand(BaseCommand):
    """Move library step position"""

    command: Literal["move-step"] = "move-step"
    step_id: str = Field(..., description="Step ID to move")
    new_position: int = Field(..., description="New position", ge=1)


class LibraryRemoveStepCommand(BaseCommand):
    """Remove library step"""

    command: Literal["remove-step"] = "remove-step"
    step_id: str = Field(..., description="Step ID to remove")


class LibraryUpdateCommand(BaseCommand):
    """Update library step"""

    command: Literal["update"] = "update"
    step_id: str = Field(..., description="Step ID to update")
    name: Optional[str] = Field(None, description="New name")
    category: Optional[str] = Field(None, description="New category")


# Union type for all library commands
LibraryCommand = Union[
    LibraryAddCommand,
    LibraryGetCommand,
    LibraryAttachCommand,
    LibraryMoveStepCommand,
    LibraryRemoveStepCommand,
    LibraryUpdateCommand,
]


# ========================================
# Master Command Union
# ========================================

VirtuosoCommand = Union[
    AssertCommand,
    InteractCommand,
    NavigateCommand,
    WindowCommand,
    DataCommand,
    DialogCommand,
    WaitCommand,
    FileCommand,
    MiscCommand,
    LibraryCommand,
]


# ========================================
# Command Group Models (for routing)
# ========================================


class StepAssertGroup(BaseModel):
    """Group model for assertion commands"""

    command_type: Literal["step-assert"] = "step-assert"
    command: AssertCommand


class StepInteractGroup(BaseModel):
    """Group model for interaction commands"""

    command_type: Literal["step-interact"] = "step-interact"
    command: InteractCommand


class StepNavigateGroup(BaseModel):
    """Group model for navigation commands"""

    command_type: Literal["step-navigate"] = "step-navigate"
    command: NavigateCommand


class StepWindowGroup(BaseModel):
    """Group model for window commands"""

    command_type: Literal["step-window"] = "step-window"
    command: WindowCommand


class StepDataGroup(BaseModel):
    """Group model for data commands"""

    command_type: Literal["step-data"] = "step-data"
    command: DataCommand


class StepDialogGroup(BaseModel):
    """Group model for dialog commands"""

    command_type: Literal["step-dialog"] = "step-dialog"
    command: DialogCommand


class StepWaitGroup(BaseModel):
    """Group model for wait commands"""

    command_type: Literal["step-wait"] = "step-wait"
    command: WaitCommand


class StepFileGroup(BaseModel):
    """Group model for file commands"""

    command_type: Literal["step-file"] = "step-file"
    command: FileCommand


class StepMiscGroup(BaseModel):
    """Group model for misc commands"""

    command_type: Literal["step-misc"] = "step-misc"
    command: MiscCommand


class LibraryGroup(BaseModel):
    """Group model for library commands"""

    command_type: Literal["library"] = "library"
    command: LibraryCommand


# ========================================
# Simplified Test Step Models (for YAML/JSON tests)
# ========================================


class SimpleNavigateStep(BaseModel):
    """Simplified navigate step"""

    navigate: str = Field(..., description="URL to navigate to")


class SimpleClickStep(BaseModel):
    """Simplified click step"""

    click: str = Field(..., description="Selector to click")


class SimpleWriteStep(BaseModel):
    """Simplified write step"""

    write: Dict[str, str] = Field(..., description="Selector and text")

    @field_validator("write")
    def validate_write(cls, v):
        if "selector" not in v or "text" not in v:
            raise ValueError("Write step must have 'selector' and 'text' fields")
        return v


class SimpleAssertStep(BaseModel):
    """Simplified assert step"""

    assert_: str = Field(
        ..., alias="assert", description="Text or selector to assert exists"
    )


class SimpleWaitStep(BaseModel):
    """Simplified wait step"""

    wait: Union[str, int] = Field(
        ..., description="Selector to wait for or milliseconds"
    )


class SimpleStoreStep(BaseModel):
    """Simplified store step"""

    store: Dict[str, str] = Field(..., description="Store configuration")

    @field_validator("store")
    def validate_store(cls, v):
        if "selector" not in v or "as" not in v:
            raise ValueError("Store step must have 'selector' and 'as' fields")
        return v


class SimpleScrollStep(BaseModel):
    """Simplified scroll step"""

    scroll: Union[str, Dict[str, Any]] = Field(
        ..., description="Scroll target or configuration"
    )


class SimpleExecuteStep(BaseModel):
    """Simplified execute step"""

    execute: str = Field(..., description="JavaScript code to execute")


class SimpleCommentStep(BaseModel):
    """Simplified comment step"""

    comment: str = Field(..., description="Comment text")


# Union of all simplified step types
SimplifiedStep = Union[
    SimpleNavigateStep,
    SimpleClickStep,
    SimpleWriteStep,
    SimpleAssertStep,
    SimpleWaitStep,
    SimpleStoreStep,
    SimpleScrollStep,
    SimpleExecuteStep,
    SimpleCommentStep,
    Dict[str, Any],  # For other step types
]
