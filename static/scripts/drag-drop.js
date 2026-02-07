// Drag-and-drop functionality for botTaskTracker kanban board
// Using SortableJS with Datastar SSE updates

function initDragDrop() {
  const columns = ['backlog', 'in_progress', 'review', 'done'];
  
  columns.forEach(columnKey => {
    const columnEl = document.getElementById('column-' + columnKey);
    if (!columnEl) return;
    
    new Sortable(columnEl, {
      group: 'kanban',
      animation: 150,
      ghostClass: 'sortable-ghost',
      dragClass: 'sortable-drag',
      chosenClass: 'sortable-chosen',
      handle: '.task-card',
      
      // Visual feedback
      onChoose: function(evt) {
        evt.item.style.opacity = '0.5';
      },
      
      onUnchoose: function(evt) {
        evt.item.style.opacity = '1';
      },
      
      // Handle drop
      onEnd: function(evt) {
        evt.item.style.opacity = '1';
        
        const taskId = evt.item.dataset.taskId;
        const fromColumn = evt.from.id.replace('column-', '');
        const toColumn = evt.to.id.replace('column-', '');
        const newPosition = evt.newIndex;
        
        // Update if column changed OR position changed within same column
        if (fromColumn !== toColumn) {
          updateTaskColumn(taskId, toColumn, newPosition);
        } else if (evt.oldIndex !== newPosition) {
          updateTaskPosition(taskId, toColumn, newPosition);
        }
      }
    });
  });
}

// Process SSE events from the response stream
async function processSSEStream(response) {
  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';
  
  while (true) {
    const { done, value } = await reader.read();
    if (done) break;
    
    buffer += decoder.decode(value, { stream: true });
    const lines = buffer.split('\n\n');
    
    // Keep the last incomplete event in buffer
    buffer = lines.pop();
    
    for (const eventBlock of lines) {
      if (!eventBlock.trim()) continue;
      
      const eventLines = eventBlock.split('\n');
      let eventType = '';
      let eventData = '';
      
      for (const line of eventLines) {
        if (line.startsWith('event: ')) {
          eventType = line.substring(7);
        } else if (line.startsWith('data: ')) {
          eventData += line.substring(6) + '\n';
        }
      }
      
      // Process datastar-patch-elements events
      if (eventType === 'datastar-patch-elements' && eventData) {
        // Extract HTML from the data
        const dataLines = eventData.trim().split('\n');
        let html = '';
        for (const line of dataLines) {
          if (line.startsWith('elements ')) {
            html += line.substring(9) + '\n';
          } else {
            html += line + '\n';
          }
        }
        
        // Parse and update the DOM
        if (html.trim()) {
          const tempDiv = document.createElement('div');
          tempDiv.innerHTML = html.trim();
          const newElement = tempDiv.firstElementChild;
          
          if (newElement && newElement.id) {
            const oldElement = document.getElementById(newElement.id);
            if (oldElement && oldElement.parentNode) {
              oldElement.parentNode.replaceChild(newElement, oldElement);
              // Re-initialize drag-drop after DOM update
              setTimeout(initDragDrop, 50);
            }
          }
        }
      }
    }
  }
}

// Send PATCH request to update task column
async function updateTaskColumn(taskId, newColumn, newPosition) {
  try {
    const response = await fetch(`/datastar/tasks/${taskId}/column`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'text/event-stream'
      },
      body: JSON.stringify({ 
        column: newColumn,
        position: newPosition 
      })
    });
    
    if (!response.ok) {
      console.error('Failed to update task column:', response.statusText);
      window.location.reload();
      return;
    }
    
    // Process the SSE event stream
    await processSSEStream(response);
    
  } catch (error) {
    console.error('Error updating task column:', error);
    window.location.reload();
  }
}

// Send PATCH request to update task position within same column
async function updateTaskPosition(taskId, column, newPosition) {
  try {
    const response = await fetch(`/datastar/tasks/${taskId}/position`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'text/event-stream'
      },
      body: JSON.stringify({ 
        column: column,
        position: newPosition 
      })
    });
    
    if (!response.ok) {
      console.error('Failed to update task position:', response.statusText);
      window.location.reload();
      return;
    }
    
    // Process the SSE event stream
    await processSSEStream(response);
    
  } catch (error) {
    console.error('Error updating task position:', error);
    window.location.reload();
  }
}

// Initialize when DOM is ready
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', initDragDrop);
} else {
  initDragDrop();
}

// Re-initialize after Datastar updates the DOM
document.addEventListener('datastar-merge-fragments', () => {
  setTimeout(initDragDrop, 100);
});
