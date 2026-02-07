// Drag-and-drop functionality for botTaskTracker kanban board
// Using SortableJS with simple column refresh after updates

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
      delay: 150,
      delayOnTouchOnly: true,
      touchStartThreshold: 5,
      
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
          updateTaskColumn(taskId, toColumn, newPosition, fromColumn);
        } else if (evt.oldIndex !== newPosition) {
          updateTaskPosition(taskId, toColumn, newPosition);
        }
      }
    });
  });
}

// Fetch and replace a column's content
async function refreshColumn(columnKey) {
  try {
    const response = await fetch(`/columns/${columnKey}`);
    if (!response.ok) {
      console.error(`Failed to refresh column ${columnKey}:`, response.statusText);
      return;
    }
    
    const html = await response.text();
    const columnEl = document.getElementById('column-' + columnKey);
    
    if (columnEl) {
      // Replace the column content
      columnEl.innerHTML = html;
      
      // Re-initialize drag-drop for this column
      const sortableInstance = Sortable.get(columnEl);
      if (sortableInstance) {
        sortableInstance.destroy();
      }
      
      // Re-create the sortable instance
      new Sortable(columnEl, {
        group: 'kanban',
        animation: 150,
        ghostClass: 'sortable-ghost',
        dragClass: 'sortable-drag',
        chosenClass: 'sortable-chosen',
        handle: '.task-card',
        delay: 150,
        delayOnTouchOnly: true,
        touchStartThreshold: 5,
        
        onChoose: function(evt) {
          evt.item.style.opacity = '0.5';
        },
        
        onUnchoose: function(evt) {
          evt.item.style.opacity = '1';
        },
        
        onEnd: function(evt) {
          evt.item.style.opacity = '1';
          
          const taskId = evt.item.dataset.taskId;
          const fromColumn = evt.from.id.replace('column-', '');
          const toColumn = evt.to.id.replace('column-', '');
          const newPosition = evt.newIndex;
          
          if (fromColumn !== toColumn) {
            updateTaskColumn(taskId, toColumn, newPosition, fromColumn);
          } else if (evt.oldIndex !== newPosition) {
            updateTaskPosition(taskId, toColumn, newPosition);
          }
        }
      });
    }
  } catch (error) {
    console.error(`Error refreshing column ${columnKey}:`, error);
  }
}

// Send PATCH request to update task column
async function updateTaskColumn(taskId, newColumn, newPosition, oldColumn) {
  try {
    const response = await fetch(`/datastar/tasks/${taskId}/column`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
        'X-Client-Nonce': window.clientNonce
      },
      body: JSON.stringify({ 
        column: newColumn,
        position: newPosition 
      })
    });
    
    if (!response.ok) {
      console.error('Failed to update task column:', response.statusText);
      // Don't refresh - let user see the error
      return;
    }
    
    // Success - don't refresh, card is already in place visually
    console.log('Column update successful');
    
  } catch (error) {
    console.error('Error updating task column:', error);
    // Refresh both columns to revert the UI change
    await Promise.all([
      refreshColumn(oldColumn),
      refreshColumn(newColumn)
    ]);
  }
}

// Send PATCH request to update task position within same column
async function updateTaskPosition(taskId, column, newPosition) {
  try {
    const response = await fetch(`/datastar/tasks/${taskId}/position`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
        'X-Client-Nonce': window.clientNonce
      },
      body: JSON.stringify({ 
        column: column,
        position: newPosition 
      })
    });
    
    if (!response.ok) {
      console.error('Failed to update task position:', response.statusText);
      // Don't refresh - let user see the error
      return;
    }
    
    // Success - don't refresh, card is already in place visually
    console.log('Position update successful');
    
  } catch (error) {
    console.error('Error updating task position:', error);
    // Refresh column to revert the UI change
    await refreshColumn(column);
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
